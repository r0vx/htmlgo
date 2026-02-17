/*

## htmlgo

Type safe and modularize way to generate html on server side.
Download the package with `go get -v github.com/r0vx/htmlgo` and import the package with `.` gives you simpler code:

	import (
		. "github.com/r0vx/htmlgo"
	)

also checkout full API documentation at: https://godoc.org/github.com/r0vx/htmlgo

*/
package htmlgo

import (
	"context"
)

type HTMLComponent interface {
	MarshalHTML(ctx context.Context, buf *[]byte) error
}

type ComponentFunc func(ctx context.Context, buf *[]byte) error

func (f ComponentFunc) MarshalHTML(ctx context.Context, buf *[]byte) error {
	return f(ctx, buf)
}

type MutableAttrHTMLComponent interface {
	HTMLComponent
	SetAttr(k string, v interface{})
}
