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

## README 是自动生成的 —— 不要手改

`modd.conf` 在任何 `.go` 改动后执行 `godoc2readme . > ./README.md`,再跑 `go test -v ./...`。所以:

- **README.md 由 godoc 生成**:内容 = `api.go` 的包注释 + `example_test.go` 里的 `Example*` 函数(含 `// Output:`)。要改 README,改这两处源,**不要直接编辑 README.md**(会被覆盖)。
- `godoc2readme` 本地未安装;若需本地重新生成需先 `go install` 该工具,否则 `modd` 的 prep 步骤会失败。
- 新增对外用法示例时,写成新的 `Example*` 函数,它同时是测试 + 文档。

## 架构(跨文件的大局)

整个库围绕**单接口 + 单 builder**:

```go
// api.go —— 唯一核心接口
type HTMLComponent interface {
    MarshalHTML(ctx context.Context, buf *[]byte) error
}
```

**1. Append-mode 缓冲是定义性设计。** 整棵组件树共享同一个 `*[]byte`,递归直接 `*buf = append(*buf, ...)`。缓冲只在顶层入口 `Fprint` / `MustString`(`utils.go`)分配一次(`make([]byte, 0, 4096)`)。这是相对上游 qor5/htmlgo 的 breaking change(旧签名 `MarshalHTML(ctx) ([]byte, error)`)。**任何新实现 `HTMLComponent` 的代码都必须 append 进传入的 buf,绝不返回或分配中间 `[]byte`。**

**2. 一切汇聚到 `HTMLTagBuilder`(`tag.go`)。** `elements.go` 里约 120 个构造器(`Div`/`Span`/`A`/`H1`…)全是 `Tag(name).Children(...)` 之类的一行包装,没有各自的渲染逻辑。真正的序列化 + 属性类型分发**只存在于** `HTMLTagBuilder.MarshalHTML`。改渲染行为基本只动这一个方法。

**3. Builder 是可变 + 链式的(`return b`),不是不可变。** 这是本库刻意的 fluent idiom(与全局"immutability"规则相反,此处以库的既定风格为准)——不要把它"修正"成返回副本。

**4. 两条转义路径,别混用:**
- 正文文本:`Text()` / `Textf()`(`utils.go`)→ `html.EscapeString`。
- 属性值:`appendEscapeAttr`(`tag.go`)→ 单引号包裹,**只转义 `'`→`&#39;` 和 `&`→`&amp;`**(`< > "` 在单引号属性里安全,故不转义)。
- `RawHTML`(`utils.go`)是逃生舱:原样输出,不转义(`Script`/`Style` 的内容就走这里)。
- 新增任何属性输出都要经过 `appendEscapeAttr`,不要手拼。

**5. 属性值语义(`MarshalHTML` 里的 type switch):** `bool` true → 裸属性、false → 省略;空字符串 → 省略;数值 → `strconv`;其余类型 → JSON(`JSONString`,**用 bytedance/sonic 而非 encoding/json**)。

**6. 条件渲染(`if.go`):** `If(cond, comps...)` 立即求值;`Iff(cond, func() HTMLComponent)` 惰性——当 body 可能 nil-panic 或开销大时用 `Iff`。两者都支持 `.ElseIf().Else()` 链。

**7. 输出格式:** 每个标签前带 `\n`(append 设计的产物,`\n<tag`),所以测试期望串都以 `\n` 开头——新增渲染测试时注意对齐这个前缀。

## 性能基线

设计文档(`docs/plans/2026-02-17-performance-optimization-design.md`)称分配降到 O(1),但实测分配数仍随组件数线性增长:buffer 已 O(1),但组件树构造仍按节点分配——`new(HTMLTagBuilder)` + attrs/children 切片、`MarshalHTML` 里的 `strings.Join`(class/styles)、`Text` 的 `html.EscapeString`。改性能时优先看构造层而非 buffer。(`Class()` 已用零分配的 `strings.SplitSeq` 替代 `strings.Split`。)
