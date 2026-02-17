package htmlgo_test

import (
	"context"
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
		expected: `
<div>
<div>Hello</div>
</div>
`,
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
		expected: `
<div>
<div class='menu' id='menu-id'>Hello</div>
</div>
`,
	},
	{
		name: "escape 1",
		tag: Div(
			Div().Text("Hello").
				Attr("class", "menu",
					"id", "the><&\"'-menu",
					"style"),
		),
		expected: `
<div>
<div class='menu' id='the><&"&#39;-menu'>Hello</div>
</div>
`,
	},
}

func TestHtmlTag(t *testing.T) {
	for _, c := range htmltagCases {
		buf := make([]byte, 0, 1024)
		err := c.tag.MarshalHTML(context.TODO(), &buf)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		got := string(buf)
		if got != c.expected {
			t.Errorf("%s:\n  expected: %q\n       got: %q", c.name, c.expected, got)
		}
	}
}
