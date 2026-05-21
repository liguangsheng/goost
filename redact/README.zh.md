# redact

面向日志行的字符串脱敏。每个辅助函数都会保留足够的原文片段，方便调试。

```go
redact.Mask("hello", 1, 1)     // "h***o"
redact.Email("alice@example.com") // "a****@example.com"
redact.Phone("13800138000")       // "138****8000"
redact.Token("abcd12345678wxyz")  // "abcd****wxyz"
```

用于结构化日志：

```go
zap.L().Info("auth",
    redact.ZapString("token", token, 4, 4),
)

slog.Info("auth",
    redact.SlogString("token", token, 4, 4),
)
```
