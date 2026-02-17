package htmlgo

import (
	"context"
	"fmt"
	"html"
	"io"
)

type RawHTML string

func (s RawHTML) MarshalHTML(ctx context.Context, buf *[]byte) error {
	*buf = append(*buf, s...)
	return nil
}

func Text(text string) (r HTMLComponent) {
	return RawHTML(html.EscapeString(text))
}

func Textf(format string, a ...interface{}) (r HTMLComponent) {
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

// Fprint 将组件渲染后写入 w，整棵树共享一个 buf
func Fprint(w io.Writer, root HTMLComponent, ctx context.Context) error {
	if root == nil {
		return nil
	}
	buf := make([]byte, 0, 4096)
	if err := root.MarshalHTML(ctx, &buf); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}

// MustString 将组件渲染为字符串，panic on error
func MustString(root HTMLComponent, ctx context.Context) string {
	buf := make([]byte, 0, 4096)
	if err := root.MarshalHTML(ctx, &buf); err != nil {
		panic(err)
	}
	return string(buf)
}
