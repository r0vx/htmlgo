# CLAUDE.md
@~/.claude/skills/golang-modern-idioms/SKILL.md
This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

`github.com/r0vx/htmlgo` —— 服务端类型安全的 HTML 生成库(builder 模式)。是 `qor5/htmlgo`(原 `theplant/htmlgo`)的 fork,改了模块路径并把渲染重写为 append-mode 缓冲。是 `r0vx/admin`、`r0vx/x` 等下游仓库的底层依赖,API 与 qor5 保持兼容,**唯一的 breaking change 是 `MarshalHTML` 的签名**(见下)。

## 常用命令

```bash
go test ./...                          # 全部测试
go test -v ./...                       # 带输出(Example 函数会校验 Output 注释)
go test -run TestHtmlTag               # 单个测试(用例是表驱动,集中在一个函数内,无子测试)
go test -bench=. -benchmem -run=^$     # 基准(基准函数在 bench_test.go)
go build ./... && go vet ./...         # 构建 + 静态检查
```

## README 与 example_test.go 同源 —— 改 Example 后必须同步 README

README 内容 = `api.go` 包注释 + `example_test.go` 的 `Example*` 函数(含 `// Output:`)。历史上由 `godoc2readme` 生成(见 `modd.conf`),**但该工具用了已删除的 `godoc.CommandLine`,新版 Go 编译不过,生成链已死**。现状:

- README.md 现在是**手工/脚本同步维护**的:改了 `Example*` 函数或其 Output 后,把对应代码块和 `//Output:` 区同步进 README(2026-06 起如此)。
- `modd.conf` 里的 `godoc2readme` prep 步骤会失败,本地跑 `modd` 时忽略该步。
- 新增对外用法示例时,仍写成新的 `Example*` 函数(测试 + 文档),并在 README 末尾按既有风格补一节。

## 架构(跨文件的大局)

整个库围绕**单接口 + 单 builder**:

```go
// api.go —— 唯一核心接口
type HTMLComponent interface {
    MarshalHTML(ctx context.Context, buf *[]byte) error
}
```

**1. Append-mode 缓冲是定义性设计。** 整棵组件树共享同一个 `*[]byte`,递归直接 `*buf = append(*buf, ...)`。顶层入口 `Fprint` / `MustString`(`utils.go`)从 `sync.Pool` 取缓冲(>1MB 不回池),稳态下渲染零 buffer 分配。这是相对上游 qor5/htmlgo 的 breaking change(旧签名 `MarshalHTML(ctx) ([]byte, error)`)。**任何新实现 `HTMLComponent` 的代码都必须 append 进传入的 buf,绝不返回或分配中间 `[]byte`。**

**2. 一切汇聚到 `HTMLTagBuilder`(`tag.go`)。** `elements.go` 里约 120 个构造器(`Div`/`Span`/`A`/`H1`…)全是 `Tag(name).Children(...)` 之类的一行包装,没有各自的渲染逻辑。真正的序列化 + 属性类型分发**只存在于** `HTMLTagBuilder.MarshalHTML` + `appendAttr`。改渲染行为基本只动这两处。

**3. Builder 是可变 + 链式的(`return b`),不是不可变。** 这是本库刻意的 fluent idiom(与全局"immutability"规则相反,此处以库的既定风格为准)——不要把它"修正"成返回副本。**但不可值拷贝 builder**(`classNames` 初始指向内联 `classBuf`,值拷贝会共享底层数组)。

**4. `MarshalHTML` 是只读的。** class/style 的 join、`Text()` 正文的转义全部延迟到渲染期直写 buf,渲染不变异 builder,重复渲染幂等(有测试固化)。

**5. 两条转义路径,别混用:**
- 正文文本:`Text()` / `Textf()`(`utils.go`)→ 存原文,渲染时 `appendEscapeText` 直接转义进 buf(规则与 `html.EscapeString` 一致,五字符)。`HTMLTagBuilder.Text()` 走 builder 的 `text` 字段快路径,零装箱。
- 属性值:`appendEscapeAttr`(`tag.go`,泛型 `string|[]byte`)→ 单引号包裹,**只转义 `'`→`&#39;` 和 `&`→`&amp;`**(`< > "` 在单引号属性里安全,故不转义)。
- `RawHTML`(`utils.go`)是逃生舱:原样输出,不转义(`Script`/`Style` 的内容就走这里)。
- 新增任何属性输出都要经过 `appendEscapeAttr`,不要手拼。

**6. 属性值语义(`appendAttr` 里的 type switch):** `bool` true → 裸属性、false → 省略;空字符串 → 省略;数值 → `strconv.Append*` 直写 buf;其余类型 → JSON(sonic 的 `[]byte` 直接转义进 buf,**不要换回 encoding/json**)。

**7. 条件渲染(`if.go`):** `If(cond, comps...)` 立即求值;`Iff(cond, func() HTMLComponent)` 惰性——当 body 可能 nil-panic 或开销大时用 `Iff`。两者都支持 `.ElseIf().Else()` 链。

**8. `Cached(comp)`(`utils.go`):** 包装 ctx 无关的静态子树(导航/footer),首次渲染后缓存字节、之后纯拷贝(整页回放 ~390ns/0 allocs)。输出依赖 ctx 或可变状态的组件不可用。

**9. 输出格式是紧凑的(无结构性换行)。** 标签间不输出 `\n`(HTML 是给浏览器的,非给人读;qor5 上游的 `\n<tag` 格式已在 v0.3.0 移除)。内容自带的换行(`RawHTML`/`Script`/`Style` 里的)原样保留。注意:行内元素间由换行产生的「隐式空格」不复存在,间距一律交给 CSS。

## 性能基线(2026-06 极致优化后)

渲染路径已无浪费;剩余分配全在**构造层**,且已被基准逐项压过:builder 本体(144B size class,**结构体大小比分配次数对耗时影响更大**——曾试过内联 attrs 数组,因结构体膨胀到 296B 反而整体变慢而回退)、call-site 变参切片、`any`/接口装箱。当前 Complex 基准 ~73µs/1229 allocs(优化前 155µs/4904)。结构体字段经过逐字节调优(`styles` 用指针懒分配省 16B 压进 144 class),**给 `HTMLTagBuilder` 加字段前必须跑基准**。再往下只有改 API(arena/池化 builder)或下游用 `Cached` 缓存静态子树。
