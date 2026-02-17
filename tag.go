package htmlgo

import (
	"context"
	"github.com/bytedance/sonic"
	"fmt"
	"html"
	"strconv"
	"strings"
)

type tagAttr struct {
	key   string
	value interface{}
}

type HTMLTagBuilder struct {
	tag        string
	omitEndTag bool
	attrs      []*tagAttr
	styles     []string
	classNames []string
	children   []HTMLComponent
}

func Tag(tag string) (r *HTMLTagBuilder) {
	r = &HTMLTagBuilder{}

	if r.attrs == nil {
		r.attrs = []*tagAttr{}
	}

	r.Tag(tag)

	return
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
	b.Children(Text(v))
	return b
}

func (b *HTMLTagBuilder) Children(comps ...HTMLComponent) (r *HTMLTagBuilder) {
	b.children = comps
	return b
}

func (b *HTMLTagBuilder) SetAttr(k string, v interface{}) {
	for _, at := range b.attrs {
		if at.key == k {
			at.value = v
			return
		}
	}
	b.attrs = append(b.attrs, &tagAttr{k, v})
}

func (b *HTMLTagBuilder) Attr(vs ...interface{}) (r *HTMLTagBuilder) {
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

func (b *HTMLTagBuilder) AttrIf(key, value interface{}, add bool) (r *HTMLTagBuilder) {
	if !add {
		return b
	}

	return b.Attr(key, value)
}

func (b *HTMLTagBuilder) Class(names ...string) (r *HTMLTagBuilder) {
	b.addClass(names...)
	return b
}

func (b *HTMLTagBuilder) addClass(names ...string) (r *HTMLTagBuilder) {
	for _, n := range names {
		ins := strings.Split(n, " ")
		for _, in := range ins {
			tin := strings.TrimSpace(in)
			if len(tin) > 0 {
				b.classNames = append(b.classNames, tin)
			}
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
		b.Attr(fmt.Sprintf("data-%s", vs[i]), vs[i+1])
	}
	return b
}

func (b *HTMLTagBuilder) Id(v string) (r *HTMLTagBuilder) {
	b.Attr("id", v)
	return b
}

func (b *HTMLTagBuilder) Href(v string) (r *HTMLTagBuilder) {
	b.Attr("href", v)
	return b
}

func (b *HTMLTagBuilder) Rel(v string) (r *HTMLTagBuilder) {
	b.Attr("rel", v)
	return b
}

func (b *HTMLTagBuilder) Title(v string) (r *HTMLTagBuilder) {
	b.Attr("title", html.EscapeString(v))
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
	b.Attr("role", v)
	return b
}

func (b *HTMLTagBuilder) Alt(v string) (r *HTMLTagBuilder) {
	b.Attr("alt", v)
	return b
}

func (b *HTMLTagBuilder) Target(v string) (r *HTMLTagBuilder) {
	b.Attr("target", v)
	return b
}

func (b *HTMLTagBuilder) Name(v string) (r *HTMLTagBuilder) {
	b.Attr("name", v)
	return b
}

func (b *HTMLTagBuilder) Value(v string) (r *HTMLTagBuilder) {
	b.Attr("value", v)
	return b
}

func (b *HTMLTagBuilder) For(v string) (r *HTMLTagBuilder) {
	b.Attr("for", v)
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
		b.styles = append(b.styles, v)
	}

	return b
}

func (b *HTMLTagBuilder) Type(v string) (r *HTMLTagBuilder) {
	b.Attr("type", v)
	return b
}

func (b *HTMLTagBuilder) Placeholder(v string) (r *HTMLTagBuilder) {
	b.Attr("placeholder", v)
	return b
}

func (b *HTMLTagBuilder) Src(v string) (r *HTMLTagBuilder) {
	b.Attr("src", v)
	return b
}

func (b *HTMLTagBuilder) Property(v string) (r *HTMLTagBuilder) {
	b.Attr("property", v)
	return b
}

func (b *HTMLTagBuilder) Action(v string) (r *HTMLTagBuilder) {
	b.Attr("action", v)
	return b
}

func (b *HTMLTagBuilder) Method(v string) (r *HTMLTagBuilder) {
	b.Attr("method", v)
	return b
}

func (b *HTMLTagBuilder) Content(v string) (r *HTMLTagBuilder) {
	b.Attr("content", v)
	return b
}

func (b *HTMLTagBuilder) Charset(v string) (r *HTMLTagBuilder) {
	b.Attr("charset", v)
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
	b.children = append(b.children, c...)
	return b
}

func (b *HTMLTagBuilder) PrependChildren(c ...HTMLComponent) (r *HTMLTagBuilder) {
	b.children = append(c, b.children...)
	return b
}

func (b *HTMLTagBuilder) MarshalHTML(ctx context.Context, buf *[]byte) error {
	class := strings.TrimSpace(strings.Join(b.classNames, " "))
	if len(class) > 0 {
		b.Attr("class", class)
	}

	styles := strings.TrimSpace(strings.Join(b.styles, "; "))
	if len(styles) > 0 {
		b.Attr("style", styles+";")
	}

	// 写开标签: \n<tag
	*buf = append(*buf, '\n', '<')
	*buf = append(*buf, b.tag...)

	// 写属性
	for _, at := range b.attrs {
		var val string
		var isBool bool
		var boolVal bool
		switch v := at.value.(type) {
		case string:
			val = v
		case []byte:
			val = string(v)
		case []rune:
			val = string(v)
		case int:
			val = strconv.Itoa(v)
		case int8:
			val = strconv.FormatInt(int64(v), 10)
		case int16:
			val = strconv.FormatInt(int64(v), 10)
		case int32:
			val = strconv.FormatInt(int64(v), 10)
		case int64:
			val = strconv.FormatInt(v, 10)
		case uint:
			val = strconv.FormatUint(uint64(v), 10)
		case uint8:
			val = strconv.FormatUint(uint64(v), 10)
		case uint16:
			val = strconv.FormatUint(uint64(v), 10)
		case uint32:
			val = strconv.FormatUint(uint64(v), 10)
		case uint64:
			val = strconv.FormatUint(v, 10)
		case float32:
			val = strconv.FormatFloat(float64(v), 'f', -1, 32)
		case float64:
			val = strconv.FormatFloat(v, 'f', -1, 64)
		case bool:
			boolVal = v
			isBool = true
		default:
			val = JSONString(v)
		}

		if len(val) == 0 && !isBool {
			continue
		}
		if isBool && !boolVal {
			continue
		}

		*buf = append(*buf, ' ')
		if isBool && boolVal {
			*buf = appendEscapeAttr(*buf, at.key)
		} else {
			*buf = appendEscapeAttr(*buf, at.key)
			*buf = append(*buf, '=', '\'')
			*buf = appendEscapeAttr(*buf, val)
			*buf = append(*buf, '\'')
		}
	}

	*buf = append(*buf, '>')

	if b.omitEndTag {
		*buf = append(*buf, '\n')
		return nil
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

func JSONString(v interface{}) (r string) {
	b, err := sonic.Marshal(v)
	if err != nil {
		panic(err)
	}
	r = string(b)
	return
}

// appendEscapeAttr 将 s 追加到 buf，遇到 ' 替换为 &#39;
func appendEscapeAttr(buf []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		if s[i] == '\'' {
			buf = append(buf, "&#39;"...)
		} else {
			buf = append(buf, s[i])
		}
	}
	return buf
}
