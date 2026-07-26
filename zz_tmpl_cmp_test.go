package htmlgo_test

import (
	"html/template"
	"io"
	"testing"

	. "github.com/r0vx/htmlgo"
)

type cmpRow struct{ Name, Email, Role string }

// 含需转义字符,确保两边都走转义路径(公平)
var cmpRows = func() []cmpRow {
	rows := make([]cmpRow, 50)
	for i := range rows {
		rows[i] = cmpRow{"User <b>Name</b>", "user&co@example.com", "admin"}
	}
	return rows
}()

// html/template:启动时预编译一次,基准只测 Execute(数据填充+上下文转义)
const pageTmpl = `<div class="container"><h1 class="title">Dashboard</h1>` +
	`<table class="data-table"><thead><tr><th>Name</th><th>Email</th><th>Role</th></tr></thead>` +
	`<tbody>{{range .}}<tr><td>{{.Name}}</td><td>{{.Email}}</td><td>{{.Role}}</td></tr>{{end}}</tbody></table></div>`

var compiledTmpl = template.Must(template.New("page").Parse(pageTmpl))

func BenchmarkTemplateExecute(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if err := compiledTmpl.Execute(io.Discard, cmpRows); err != nil {
			b.Fatal(err)
		}
	}
}

// htmlgo:每请求构造组件树 + 渲染(其真实用法)
func BenchmarkHtmlgoEquivalent(b *testing.B) {
	ctx := b.Context()
	b.ReportAllocs()
	for b.Loop() {
		rows := make([]HTMLComponent, len(cmpRows))
		for i, r := range cmpRows {
			rows[i] = Tr(Td(Text(r.Name)), Td(Text(r.Email)), Td(Text(r.Role)))
		}
		page := Div(
			H1("Dashboard").Class("title"),
			Table(
				Thead(Tr(Th("Name"), Th("Email"), Th("Role"))),
				Tbody(rows...),
			).Class("data-table"),
		).Class("container")
		Fprint(io.Discard, page, ctx)
	}
}

// htmlgo + Cached:整页静态时的回放(template 无对应能力)
var cachedPage = Cached(func() HTMLComponent {
	rows := make([]HTMLComponent, len(cmpRows))
	for i, r := range cmpRows {
		rows[i] = Tr(Td(Text(r.Name)), Td(Text(r.Email)), Td(Text(r.Role)))
	}
	return Div(
		H1("Dashboard").Class("title"),
		Table(Thead(Tr(Th("Name"), Th("Email"), Th("Role"))), Tbody(rows...)).Class("data-table"),
	).Class("container")
}())

func BenchmarkHtmlgoCached(b *testing.B) {
	ctx := b.Context()
	b.ReportAllocs()
	for b.Loop() {
		Fprint(io.Discard, cachedPage, ctx)
	}
}
