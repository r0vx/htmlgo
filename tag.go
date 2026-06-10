package htmlgo

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"html"
	"strconv"
	"strings"
)

type tagAttr struct {
	key   string
	value any
}

// HTMLTagBuilder 不可值拷贝:classNames 初始指向内联的 classBuf,
// 值拷贝会让两个 builder 共享同一底层数组
// 字段经基准调优:结构体大小(分配清零 + GC 扫描)比分配次数对耗时影响更大,
// 故只内联收益最高的单个 class 槽,不内联 attrs
type HTMLTagBuilder struct {
	tag        string
	text       string // Text() 的正文,渲染时直接转义写入 buf,避免装箱为子组件;非空即生效(Text("") 与无文本输出等价)
	attrs      []tagAttr
	styles     *[]string // 内联样式极少用,指针懒分配换 16B 结构体瘦身(160→144 size class)
	classNames []string
	children   []HTMLComponent
	classBuf   [1]string // classNames 的内联首块;Class() 存原文不拆分,单次调用即覆盖多数场景
	omitEndTag bool
}

func Tag(tag string) (r *HTMLTagBuilder) {
	return &HTMLTagBuilder{tag: tag}
}

func (b *HTMLTagBuilder) Tag(v string) (r *HTMLTagBuilder) {
	b.tag = v
	return b
}

func (b *HTMLTagBuilder) OmitEndTag() (r *HTMLTagBuilder) {
	b.omitEndTag = true
	return b
}

func (b *HTMLTagBuilder) Text(v string) (r *HTMLTagBuilder) {
	b.text = v
	b.children = nil
	return b
}

func (b *HTMLTagBuilder) Children(comps ...HTMLComponent) (r *HTMLTagBuilder) {
	b.children = comps
	b.text = ""
	return b
}

// findOrAddAttr 按 key 查找属性,不存在则追加一个空槽
func (b *HTMLTagBuilder) findOrAddAttr(k string) *tagAttr {
	for i := range b.attrs {
		if b.attrs[i].key == k {
			return &b.attrs[i]
		}
	}
	b.attrs = append(b.attrs, tagAttr{key: k})
	return &b.attrs[len(b.attrs)-1]
}

func (b *HTMLTagBuilder) SetAttr(k string, v any) {
	b.findOrAddAttr(k).value = v
}

// setAttrString 字符串属性 helper 的公共入口
func (b *HTMLTagBuilder) setAttrString(k, v string) {
	b.findOrAddAttr(k).value = v
}

func (b *HTMLTagBuilder) Attr(vs ...any) (r *HTMLTagBuilder) {
	if len(vs)%2 != 0 {
		vs = append(vs, "")
	}

	for i := 0; i < len(vs); i = i + 2 {
		if key, ok := vs[i].(string); ok {
			b.SetAttr(key, vs[i+1])
		} else {
			panic(fmt.Sprintf("Attr key must be string, but was %#+v", vs[i]))
		}
	}
	return b
}

func (b *HTMLTagBuilder) AttrIf(key, value any, add bool) (r *HTMLTagBuilder) {
	if !add {
		return b
	}

	return b.Attr(key, value)
}

func (b *HTMLTagBuilder) Class(names ...string) (r *HTMLTagBuilder) {
	b.addClass(names...)
	return b
}

// addClass 存原始字符串不拆分，拆分/trim 延迟到渲染期零分配完成
// 全空白的串在此过滤，保证 classNames 非空时渲染必有输出；首块用内联 classBuf
func (b *HTMLTagBuilder) addClass(names ...string) (r *HTMLTagBuilder) {
	for _, n := range names {
		if strings.TrimSpace(n) != "" {
			if b.classNames == nil {
				b.classNames = b.classBuf[:0]
			}
			b.classNames = append(b.classNames, n)
		}
	}
	return b
}

func (b *HTMLTagBuilder) ClassIf(name string, add bool) (r *HTMLTagBuilder) {
	if !add {
		return b
	}
	b.addClass(name)
	return b
}

func (b *HTMLTagBuilder) Data(vs ...string) (r *HTMLTagBuilder) {
	for i := 0; i < len(vs); i = i + 2 {
		b.setAttrString("data-"+vs[i], vs[i+1])
	}
	return b
}

func (b *HTMLTagBuilder) Id(v string) (r *HTMLTagBuilder) {
	b.setAttrString("id", v)
	return b
}

func (b *HTMLTagBuilder) Href(v string) (r *HTMLTagBuilder) {
	b.setAttrString("href", v)
	return b
}

func (b *HTMLTagBuilder) Rel(v string) (r *HTMLTagBuilder) {
	b.setAttrString("rel", v)
	return b
}

func (b *HTMLTagBuilder) Title(v string) (r *HTMLTagBuilder) {
	b.setAttrString("title", html.EscapeString(v))
	return b
}

func (b *HTMLTagBuilder) TabIndex(v int) (r *HTMLTagBuilder) {
	b.Attr("tabindex", v)
	return b
}

func (b *HTMLTagBuilder) Required(v bool) (r *HTMLTagBuilder) {
	b.Attr("required", v)
	return b
}

func (b *HTMLTagBuilder) Readonly(v bool) (r *HTMLTagBuilder) {
	b.Attr("readonly", v)
	return b
}

func (b *HTMLTagBuilder) Role(v string) (r *HTMLTagBuilder) {
	b.setAttrString("role", v)
	return b
}

func (b *HTMLTagBuilder) Alt(v string) (r *HTMLTagBuilder) {
	b.setAttrString("alt", v)
	return b
}

func (b *HTMLTagBuilder) Target(v string) (r *HTMLTagBuilder) {
	b.setAttrString("target", v)
	return b
}

func (b *HTMLTagBuilder) Name(v string) (r *HTMLTagBuilder) {
	b.setAttrString("name", v)
	return b
}

func (b *HTMLTagBuilder) Value(v string) (r *HTMLTagBuilder) {
	b.setAttrString("value", v)
	return b
}

func (b *HTMLTagBuilder) For(v string) (r *HTMLTagBuilder) {
	b.setAttrString("for", v)
	return b
}

func (b *HTMLTagBuilder) Style(v string) (r *HTMLTagBuilder) {
	b.addStyle(strings.Trim(v, ";"))
	return b
}

func (b *HTMLTagBuilder) StyleIf(v string, add bool) (r *HTMLTagBuilder) {
	if !add {
		return b
	}
	b.Style(v)
	return b
}

func (b *HTMLTagBuilder) addStyle(v string) (r *HTMLTagBuilder) {
	if len(v) > 0 {
		if b.styles == nil {
			b.styles = &[]string{}
		}
		*b.styles = append(*b.styles, v)
	}

	return b
}

func (b *HTMLTagBuilder) Type(v string) (r *HTMLTagBuilder) {
	b.setAttrString("type", v)
	return b
}

func (b *HTMLTagBuilder) Placeholder(v string) (r *HTMLTagBuilder) {
	b.setAttrString("placeholder", v)
	return b
}

func (b *HTMLTagBuilder) Src(v string) (r *HTMLTagBuilder) {
	b.setAttrString("src", v)
	return b
}

func (b *HTMLTagBuilder) Property(v string) (r *HTMLTagBuilder) {
	b.setAttrString("property", v)
	return b
}

func (b *HTMLTagBuilder) Action(v string) (r *HTMLTagBuilder) {
	b.setAttrString("action", v)
	return b
}

func (b *HTMLTagBuilder) Method(v string) (r *HTMLTagBuilder) {
	b.setAttrString("method", v)
	return b
}

func (b *HTMLTagBuilder) Content(v string) (r *HTMLTagBuilder) {
	b.setAttrString("content", v)
	return b
}

func (b *HTMLTagBuilder) Charset(v string) (r *HTMLTagBuilder) {
	b.setAttrString("charset", v)
	return b
}

func (b *HTMLTagBuilder) Disabled(v bool) (r *HTMLTagBuilder) {
	b.Attr("disabled", v)
	return b
}

func (b *HTMLTagBuilder) Checked(v bool) (r *HTMLTagBuilder) {
	b.Attr("checked", v)
	return b
}

func (b *HTMLTagBuilder) AppendChildren(c ...HTMLComponent) (r *HTMLTagBuilder) {
	b.materializeText()
	b.children = append(b.children, c...)
	return b
}

func (b *HTMLTagBuilder) PrependChildren(c ...HTMLComponent) (r *HTMLTagBuilder) {
	b.materializeText()
	b.children = append(c, b.children...)
	return b
}

// materializeText 把 Text() 快路径的正文转回子组件，保持与 children 的顺序语义
// Text("") 无可见输出，不必物化
func (b *HTMLTagBuilder) materializeText() {
	if len(b.text) > 0 {
		b.children = []HTMLComponent{textComponent(b.text)}
		b.text = ""
	}
}

func (b *HTMLTagBuilder) MarshalHTML(ctx context.Context, buf *[]byte) error {
	// class/style 延迟到此处直接写入 buf，渲染不变异 builder
	classPending := len(b.classNames) > 0

	var styleVal string
	if b.styles != nil && len(*b.styles) > 0 {
		styleVal = strings.TrimSpace(strings.Join(*b.styles, "; "))
	}
	stylePending := len(styleVal) > 0

	// 写开标签: \n<tag
	*buf = append(*buf, '\n', '<')
	*buf = append(*buf, b.tag...)

	// 写属性；手动 Attr 设置过的 class/style 在原位置被 Class()/Style() 的值覆盖
	for i := range b.attrs {
		at := &b.attrs[i]
		if classPending && at.key == "class" {
			b.appendClassAttr(buf)
			classPending = false
			continue
		}
		if stylePending && at.key == "style" {
			appendStyleAttr(buf, styleVal)
			stylePending = false
			continue
		}
		appendAttr(buf, at)
	}

	if classPending {
		b.appendClassAttr(buf)
	}
	if stylePending {
		appendStyleAttr(buf, styleVal)
	}

	*buf = append(*buf, '>')

	if b.omitEndTag {
		*buf = append(*buf, '\n')
		return nil
	}

	// Text() 快路径：正文直接转义写入，无子组件装箱
	if len(b.text) > 0 {
		*buf = appendEscapeText(*buf, b.text)
	}

	// 递归写子组件
	for _, c := range b.children {
		if c == nil {
			continue
		}
		if err := c.MarshalHTML(ctx, buf); err != nil {
			return err
		}
	}

	// 写闭标签: </tag>\n
	*buf = append(*buf, '<', '/')
	*buf = append(*buf, b.tag...)
	*buf = append(*buf, '>', '\n')
	return nil
}

func JSONString(v any) (r string) {
	b, err := sonic.Marshal(v)
	if err != nil {
		panic(err)
	}
	r = string(b)
	return
}

// appendClassAttr 将 classNames 拆分、trim 后以单空格连接写为 class 属性
// classNames 存的是 Class() 的原始入参，此处用 SplitSeq 零分配拆分
func (b *HTMLTagBuilder) appendClassAttr(buf *[]byte) {
	*buf = append(*buf, " class='"...)
	first := true
	for _, raw := range b.classNames {
		for in := range strings.SplitSeq(raw, " ") {
			n := strings.TrimSpace(in)
			if n == "" {
				continue
			}
			if !first {
				*buf = append(*buf, ' ')
			}
			first = false
			*buf = appendEscapeAttr(*buf, n)
		}
	}
	*buf = append(*buf, '\'')
}

// appendStyleAttr 写 style 属性，值已 join + trim，此处补尾分号
func appendStyleAttr(buf *[]byte, styleVal string) {
	*buf = append(*buf, " style='"...)
	*buf = appendEscapeAttr(*buf, styleVal)
	*buf = append(*buf, ';', '\'')
}

// appendAttrKey 写属性名前导空格 + 转义后的 key
func appendAttrKey(buf *[]byte, key string) {
	*buf = append(*buf, ' ')
	*buf = appendEscapeAttr(*buf, key)
}

// appendAttrStr 写完整的 key='val'（val 非空时）
func appendAttrStr[T ~string | ~[]byte](buf *[]byte, key string, val T) {
	if len(val) == 0 {
		return
	}
	appendAttrKey(buf, key)
	*buf = append(*buf, '=', '\'')
	*buf = appendEscapeAttr(*buf, val)
	*buf = append(*buf, '\'')
}

// appendAttrInt / appendAttrUint / appendAttrFloat 数值属性直写
// 数值输出不含 ' 和 &，无需转义
func appendAttrInt(buf *[]byte, key string, v int64) {
	appendAttrKey(buf, key)
	*buf = append(*buf, '=', '\'')
	*buf = strconv.AppendInt(*buf, v, 10)
	*buf = append(*buf, '\'')
}

func appendAttrUint(buf *[]byte, key string, v uint64) {
	appendAttrKey(buf, key)
	*buf = append(*buf, '=', '\'')
	*buf = strconv.AppendUint(*buf, v, 10)
	*buf = append(*buf, '\'')
}

func appendAttrFloat(buf *[]byte, key string, v float64, bits int) {
	appendAttrKey(buf, key)
	*buf = append(*buf, '=', '\'')
	*buf = strconv.AppendFloat(*buf, v, 'f', -1, bits)
	*buf = append(*buf, '\'')
}

// appendAttr 按值类型直写属性，数值经 strconv.Append* 零中间字符串
// 语义: bool true → 裸属性、false → 省略；空字符串 → 省略；其余类型 → JSON(sonic)
func appendAttr(buf *[]byte, at *tagAttr) {
	switch v := at.value.(type) {
	case string:
		appendAttrStr(buf, at.key, v)
	case []byte:
		appendAttrStr(buf, at.key, v)
	case []rune:
		appendAttrStr(buf, at.key, string(v))
	case int:
		appendAttrInt(buf, at.key, int64(v))
	case int8:
		appendAttrInt(buf, at.key, int64(v))
	case int16:
		appendAttrInt(buf, at.key, int64(v))
	case int32:
		appendAttrInt(buf, at.key, int64(v))
	case int64:
		appendAttrInt(buf, at.key, v)
	case uint:
		appendAttrUint(buf, at.key, uint64(v))
	case uint8:
		appendAttrUint(buf, at.key, uint64(v))
	case uint16:
		appendAttrUint(buf, at.key, uint64(v))
	case uint32:
		appendAttrUint(buf, at.key, uint64(v))
	case uint64:
		appendAttrUint(buf, at.key, v)
	case float32:
		appendAttrFloat(buf, at.key, float64(v), 32)
	case float64:
		appendAttrFloat(buf, at.key, v, 64)
	case bool:
		if v {
			appendAttrKey(buf, at.key) // 裸属性
		}
	default:
		jb, err := sonic.Marshal(v)
		if err != nil {
			panic(err)
		}
		appendAttrStr(buf, at.key, jb)
	}
}

// appendEscapeAttr 将 s 追加到 buf，转义 HTML 属性值中的特殊字符
// 单引号属性值中需转义: ' → &#39;, & → &amp; (防止浏览器解码 HTML 实体如 &quot;)
// 无特殊字符时整段批量拷贝，不逐字节 append
func appendEscapeAttr[T ~string | ~[]byte](buf []byte, s T) []byte {
	start := 0
	for i := 0; i < len(s); i++ {
		var esc string
		switch s[i] {
		case '\'':
			esc = "&#39;"
		case '&':
			esc = "&amp;"
		default:
			continue
		}
		buf = append(buf, s[start:i]...)
		buf = append(buf, esc...)
		start = i + 1
	}
	return append(buf, s[start:]...)
}
