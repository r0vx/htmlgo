package htmlgo

import (
	"context"
	"fmt"
	"io"
	"sync"
)

type RawHTML string

func (s RawHTML) MarshalHTML(ctx context.Context, buf *[]byte) error {
	*buf = append(*buf, s...)
	return nil
}

// textComponent 持有原文，渲染时直接转义进共享 buf，避免 html.EscapeString 的中间字符串
type textComponent string

func (s textComponent) MarshalHTML(ctx context.Context, buf *[]byte) error {
	*buf = appendEscapeText(*buf, string(s))
	return nil
}

// appendEscapeText 按 html.EscapeString 的规则转义正文文本并追加到 buf
// 转义 < > & ' " 五个字符；无特殊字符时整段批量拷贝
func appendEscapeText(buf []byte, s string) []byte {
	start := 0
	for i := 0; i < len(s); i++ {
		var esc string
		switch s[i] {
		case '&':
			esc = "&amp;"
		case '\'':
			esc = "&#39;"
		case '<':
			esc = "&lt;"
		case '>':
			esc = "&gt;"
		case '"':
			esc = "&#34;"
		default:
			continue
		}
		buf = append(buf, s[start:i]...)
		buf = append(buf, esc...)
		start = i + 1
	}
	return append(buf, s[start:]...)
}

func Text(text string) (r HTMLComponent) {
	return textComponent(text)
}

func Textf(format string, a ...any) (r HTMLComponent) {
	return Text(fmt.Sprintf(format, a...))
}

type HTMLComponents []HTMLComponent

func Components(comps ...HTMLComponent) HTMLComponents {
	return HTMLComponents(comps)
}

func (hcs HTMLComponents) MarshalHTML(ctx context.Context, buf *[]byte) error {
	for _, h := range hcs {
		if h == nil {
			continue
		}
		if err := h.MarshalHTML(ctx, buf); err != nil {
			return err
		}
	}
	return nil
}

// maxPooledBufCap 超过此容量的 buffer 不回池，避免个别巨页长期占住内存
const maxPooledBufCap = 1 << 20 // 1MB

var bufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 4096)
		return &b
	},
}

func getBuf() *[]byte { return bufPool.Get().(*[]byte) }

func putBuf(p *[]byte) {
	if cap(*p) > maxPooledBufCap {
		return
	}
	*p = (*p)[:0]
	bufPool.Put(p)
}

// Fprint 将组件渲染后写入 w，整棵树共享一个池化 buf；稳态下渲染零 buffer 分配
func Fprint(w io.Writer, root HTMLComponent, ctx context.Context) error {
	if root == nil {
		return nil
	}
	p := getBuf()
	defer putBuf(p)
	if err := root.MarshalHTML(ctx, p); err != nil {
		return err
	}
	_, err := w.Write(*p)
	return err
}

// MustString 将组件渲染为字符串，panic on error
func MustString(root HTMLComponent, ctx context.Context) string {
	if root == nil {
		return ""
	}
	p := getBuf()
	defer putBuf(p)
	if err := root.MarshalHTML(ctx, p); err != nil {
		panic(err)
	}
	return string(*p)
}

// cachedComponent 见 Cached
type cachedComponent struct {
	comp HTMLComponent
	once sync.Once
	html []byte
	err  error
}

// Cached 包装一个与 ctx 无关的静态组件:首次渲染后缓存字节，之后纯拷贝输出。
// 适合导航、footer 等每请求重复构建的不变子树；被包装组件的输出若依赖
// ctx 或可变状态则不要使用。并发安全。
func Cached(comp HTMLComponent) HTMLComponent {
	return &cachedComponent{comp: comp}
}

func (c *cachedComponent) MarshalHTML(ctx context.Context, buf *[]byte) error {
	c.once.Do(func() {
		b := make([]byte, 0, 1024)
		c.err = c.comp.MarshalHTML(ctx, &b)
		c.html = b
	})
	if c.err != nil {
		return c.err
	}
	*buf = append(*buf, c.html...)
	return nil
}
