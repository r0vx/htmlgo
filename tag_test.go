package htmlgo_test

import (
	"strings"
	"sync"
	"testing"

	. "github.com/r0vx/htmlgo"
)

var htmltagCases = []struct {
	name     string
	tag      *HTMLTagBuilder
	expected string
}{
	{
		name: "case 1",
		tag: Div(
			Div().Text("Hello"),
		),
		expected: `<div><div>Hello</div></div>`,
	},
	{
		name: "case 2",
		tag: Div(
			Div().Text("Hello").
				Attr("class", "menu",
					"id", "the-menu",
					"style").
				Attr("id", "menu-id"),
		),
		expected: `<div><div class='menu' id='menu-id'>Hello</div></div>`,
	},
	{
		name: "escape 1",
		tag: Div(
			Div().Text("Hello").
				Attr("class", "menu",
					"id", "the><&\"'-menu",
					"style"),
		),
		expected: `<div><div class='menu' id='the><&amp;"&#39;-menu'>Hello</div></div>`,
	},
	{
		// 手动 Attr("class") 与 Class() 共存时,Class() 的值在原属性位置覆盖
		name: "class override keeps position",
		tag: Div().
			Attr("class", "manual", "id", "x").
			Class("a b").Class("c"),
		expected: `<div class='a b c' id='x'></div>`,
	},
	{
		name: "style joined with trailing semicolon",
		tag: Div().
			Style("color: red;").
			Style("margin: 0").
			Class("box'&"),
		expected: `<div class='box&#39;&amp;' style='color: red; margin: 0;'></div>`,
	},
	{
		// 超过内联容量(2)的属性和 class 溢出到堆,顺序与去重不变
		name: "inline buffer overflow",
		tag: Div().
			Attr("a", "1", "b", "2", "c", "3", "d", "4").
			Attr("b", "20").
			Class("c1").Class("c2").Class("c3 c4"),
		expected: `<div a='1' b='20' c='3' d='4' class='c1 c2 c3 c4'></div>`,
	},
	{
		// Text 快路径与 Append/PrependChildren 物化后的顺序语义
		name: "text materialize ordering",
		tag: Div(
			P().Text("mid").AppendChildren(Span("after")).PrependChildren(Span("before")),
		),
		expected: `<div><p><span>before</span>mid<span>after</span></p></div>`,
	},
	{
		// Text 之后 Children 整体替换;Children 之后 Text 也整体替换
		name: "text replaced by children and vice versa",
		tag: Div(
			P().Text("gone").Children(Span("kept")),
			B("gone too").Text("final"),
		),
		expected: `<div><p><span>kept</span></p><b>final</b></div>`,
	},
	{
		// 数值/bool/[]byte/JSON 属性直写路径
		name: "typed attr values",
		tag: Input("n").
			Attr("tabindex", 3).
			Attr("data-f", 1.5).
			Attr("data-b", []byte("bs")).
			Attr("data-json", struct {
				A int `json:"a"`
			}{7}).
			Disabled(true).
			Readonly(false),
		expected: `<input name='n' tabindex='3' data-f='1.5' data-b='bs' data-json='{"a":7}' disabled>`,
	},
	{
		// 免装箱通道与 Attr(any) 通道互相覆盖,后写者生效
		name: "string channel and any channel dedup",
		tag: Div().
			Href("/old").Attr("href", "/mid").Href("/new").
			Attr("id", "x").Id("y"),
		expected: `<div href='/new' id='y'></div>`,
	},
}

func TestHtmlTag(t *testing.T) {
	for _, c := range htmltagCases {
		// 渲染两次校验幂等:MarshalHTML 是只读的,不变异 builder
		for range 2 {
			buf := make([]byte, 0, 1024)
			err := c.tag.MarshalHTML(t.Context(), &buf)
			if err != nil {
				t.Fatalf("%s: %v", c.name, err)
			}
			got := string(buf)
			if got != c.expected {
				t.Errorf("%s:\n  expected: %q\n       got: %q", c.name, c.expected, got)
			}
		}
	}
}

// TestFprintPoolReuse 池化 buffer 复用后输出必须互不污染
func TestFprintPoolReuse(t *testing.T) {
	var b1, b2 strings.Builder
	if err := Fprint(&b1, Div().Class("a"), t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := Fprint(&b2, Span("x"), t.Context()); err != nil {
		t.Fatal(err)
	}
	if b1.String() != "<div class='a'></div>" {
		t.Errorf("b1 polluted: %q", b1.String())
	}
	if b2.String() != "<span>x</span>" {
		t.Errorf("b2 polluted: %q", b2.String())
	}
}

// TestCached 静态子树缓存:多次/并发渲染输出一致
func TestCached(t *testing.T) {
	nav := Cached(Nav(Li(A(Text("Home")).Href("/"))).Class("main"))
	want := MustString(Nav(Li(A(Text("Home")).Href("/"))).Class("main"), t.Context())

	for range 3 {
		if got := MustString(nav, t.Context()); got != want {
			t.Fatalf("cached output mismatch:\n  want %q\n   got %q", want, got)
		}
	}

	var wg sync.WaitGroup
	errs := make([]string, 8)
	for i := range 8 {
		wg.Go(func() {
			if got := MustString(nav, t.Context()); got != want {
				errs[i] = got
			}
		})
	}
	wg.Wait()
	for i, e := range errs {
		if e != "" {
			t.Errorf("goroutine %d mismatch: %q", i, e)
		}
	}
}
