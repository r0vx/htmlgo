package htmlgo_test

import (
	"context"
	"strings"
	"testing"

	. "github.com/r0vx/htmlgo"
)

// TestTextNeutralizesVueInterpolation 验证 Text 默认中和 `{{`（Vue 插值起始），
// 防止用户数据被浏览器端当模板执行；单个 `{` 原样保留；Interp 原样透传插值。
func TestTextNeutralizesVueInterpolation(t *testing.T) {
	render := func(c HTMLComponent) string {
		return MustString(c, context.Background())
	}

	// 用户数据里的 {{expr}} 必须被中和：首个 `{` 转实体，不再形成 `{{` 分隔符
	got := render(Text("{{constructor.constructor('alert(1)')()}}"))
	if strings.Contains(got, "{{") {
		t.Fatalf("Text 未中和 `{{`，仍可被 Vue 当插值执行: %q", got)
	}
	if !strings.Contains(got, "&#123;") {
		t.Fatalf("期望首个 `{` 转为 &#123;，得到: %q", got)
	}

	// 单个 `{` 不转义（减少输出噪声）
	if got := render(Text("a {b} c")); got != "a {b} c" {
		t.Fatalf("单个 `{` 不应转义, 得到: %q", got)
	}

	// 结尾单个 `{` 不转义
	if got := render(Text("x{")); got != "x{" {
		t.Fatalf("结尾单个 `{` 不应转义, 得到: %q", got)
	}

	// Interp 原样透传（框架受信插值），不转义
	if got := render(Interp("{{locals.count}}")); got != "{{locals.count}}" {
		t.Fatalf("Interp 应原样输出插值, 得到: %q", got)
	}

	// 其余 HTML 特殊字符仍照常转义
	if got := render(Text("<b>&\"")); got != "&lt;b&gt;&amp;&#34;" {
		t.Fatalf("HTML 转义回归: %q", got)
	}
}
