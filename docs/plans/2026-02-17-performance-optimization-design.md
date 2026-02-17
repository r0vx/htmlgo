# htmlgo 性能优化设计

## 背景

当前 `MarshalHTML` 返回 `[]byte`，导致深层嵌套时产生大量中间分配。

Benchmark 基线（Apple M2 Max）：

| 场景 | 耗时 | 内存 | 分配次数 |
|------|------|------|---------|
| Simple（3组件） | 1.27 us | 1.3 KB | 42 |
| Medium（~70组件） | 30 us | 33 KB | 901 |
| Complex（~300组件） | 445 us | 566 KB | 13,417 |

## 方案

Breaking change，将 `MarshalHTML` 签名改为 `*[]byte` append 模式（参考 sonic 的 `EncodeInto`）。

## 接口变更

```go
// 旧
type HTMLComponent interface {
    MarshalHTML(ctx context.Context) ([]byte, error)
}

// 新
type HTMLComponent interface {
    MarshalHTML(ctx context.Context, buf *[]byte) error
}

type ComponentFunc func(ctx context.Context, buf *[]byte) error
```

## 核心改动

### 1. HTMLTagBuilder.MarshalHTML

- 消除所有 `fmt.Sprintf`，改用 `*buf = append(*buf, ...)` 直接拼接
- 子组件递归共享同一个 `buf`
- 移除 `bufPool`（不再需要）

### 2. appendEscapeAttr

替代 `escapeAttr` + `strings.Replace`，逐字节扫描直接 append 到 buf。

### 3. 辅助类型

- `RawHTML.MarshalHTML`: `*buf = append(*buf, s...)`
- `HTMLComponents.MarshalHTML`: 遍历子组件，共享 buf
- `IfBuilder` / `IfFuncBuilder`: 签名适配

### 4. 顶层入口

- `Fprint`: 唯一做 `make([]byte, 0, 4096)` 的地方，最后一次 `w.Write(buf)`
- `MustString`: 同理，最后 `string(buf)`

## 预期收益

- 内存分配从 O(组件数) 降到 O(1)
- 消除 `fmt.Sprintf` 反射开销
- 消除 `string(b)` 冗余拷贝
- 预期整体 2-3x 提升

## 影响范围

约 23 个仓库依赖此库，核心是 qor5/admin、qor5/x。所有实现 `HTMLComponent` 接口的代码需适配新签名。
