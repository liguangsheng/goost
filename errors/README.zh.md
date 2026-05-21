# errors

标准库 `errors` 之上的薄封装：在 wrap 位置捕获 stack trace。完全兼容
`errors.Is` / `errors.As` / `errors.Unwrap` 和
`fmt.Errorf("...: %w", err)`。

```go
import "github.com/liguangsheng/goost/errors"

func read() error {
    f, err := os.Open("x")
    if err != nil {
        return errors.Wrap(err, "open config")
    }
    defer f.Close()
    return nil
}

// 调用方：
if err := read(); err != nil {
    log.Printf("%+v", err) // 打印消息和 stack frames
}
```

| 辅助函数 | 行为 |
| --- | --- |
| `New(msg)` | 类似 `errors.New`，并在调用点附带 stack。 |
| `Errorf(fmt, args...)` | 类似 `fmt.Errorf`；保留 `%w`；捕获 stack。 |
| `WithStack(err)` | 附加 stack；如果已经有 stack 则不重复附加。 |
| `Wrap(err, msg)` | 添加注释并捕获新的 stack。输入 nil 输出 nil。 |
| `Wrapf(err, fmt, args...)` | `Wrap` 的格式化版本。 |
| `StackTrace(err)` | 返回捕获的 PCs；没有则返回 nil。 |
| `FormatStack(err)` | 把 stack 渲染成多行字符串。 |
