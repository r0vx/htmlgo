package htmlgo_test

import (
	"context"
	"io"
	"testing"

	. "github.com/r0vx/htmlgo"
)

// 模拟一个简单页面：少量组件
func simpleComponent() HTMLComponent {
	return Div(
		H1("Hello World"),
		P(Text("This is a paragraph.")),
		A(Text("Click me")).Href("/link"),
	)
}

// 模拟一个中等复杂度的页面
func mediumComponent() HTMLComponent {
	rows := make([]HTMLComponent, 20)
	for i := 0; i < 20; i++ {
		rows[i] = Tr(
			Td(Text("Cell 1")),
			Td(Text("Cell 2")),
			Td(Text("Cell 3")),
		)
	}
	return Div(
		H1("Dashboard").Class("title"),
		Nav(
			Ul(
				Li(A(Text("Home")).Href("/")),
				Li(A(Text("About")).Href("/about")),
				Li(A(Text("Contact")).Href("/contact")),
			),
		).Class("nav"),
		Table(
			Thead(
				Tr(
					Th("Name"),
					Th("Email"),
					Th("Role"),
				),
			),
			Tbody(rows...),
		).Class("data-table"),
	).Class("container")
}

// 模拟一个复杂的真实页面：嵌套深、组件多
func complexComponent() HTMLComponent {
	cards := make([]HTMLComponent, 50)
	for i := 0; i < 50; i++ {
		cards[i] = Div(
			Div(
				Img("/img/photo.jpg").Alt("photo").Class("card-img"),
			).Class("card-header"),
			Div(
				H3("Card Title").Class("card-title"),
				P(Text("Some description text that goes in the card body.")).Class("card-text"),
				Div(
					Span("Tag1").Class("badge"),
					Span("Tag2").Class("badge"),
					Span("Tag3").Class("badge"),
				).Class("tags"),
			).Class("card-body"),
			Div(
				A(Text("Read more")).Href("/detail").Class("btn btn-primary"),
				Button("Save").Class("btn btn-secondary"),
			).Class("card-footer"),
		).Class("card")
	}

	return HTML(
		Head(
			Meta().Charset("utf-8"),
			Title("Complex Page"),
			Link("/css/style.css").Rel("stylesheet"),
			Link("/css/theme.css").Rel("stylesheet"),
		),
		Body(
			Header(
				Nav(
					Div(
						A(Text("Brand")).Href("/").Class("brand"),
						Ul(
							Li(A(Text("Home")).Href("/")),
							Li(A(Text("Products")).Href("/products")),
							Li(A(Text("About")).Href("/about")),
							Li(A(Text("Contact")).Href("/contact")),
						).Class("nav-links"),
					).Class("nav-inner"),
				).Class("main-nav"),
			).Class("header"),
			Main(
				Section(
					H1("Welcome").Class("hero-title"),
					P(Text("This is the hero section of the page.")).Class("hero-text"),
				).Class("hero"),
				Section(cards...).Class("card-grid"),
			).Class("content"),
			Footer(
				Div(
					P(Text("Copyright 2024")),
					Ul(
						Li(A(Text("Privacy")).Href("/privacy")),
						Li(A(Text("Terms")).Href("/terms")),
					),
				).Class("footer-inner"),
			).Class("footer"),
		),
	)
}

func BenchmarkSimple(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		comp := simpleComponent()
		buf := make([]byte, 0, 1024)
		if err := comp.MarshalHTML(ctx, &buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMedium(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		comp := mediumComponent()
		buf := make([]byte, 0, 4096)
		if err := comp.MarshalHTML(ctx, &buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkComplex(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		comp := complexComponent()
		buf := make([]byte, 0, 16384)
		if err := comp.MarshalHTML(ctx, &buf); err != nil {
			b.Fatal(err)
		}
	}
}

// 单独测 Fprint 路径（模拟 HTTP handler 场景）
func BenchmarkComplexFprint(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		comp := complexComponent()
		if err := Fprint(io.Discard, comp, ctx); err != nil {
			b.Fatal(err)
		}
	}
}
