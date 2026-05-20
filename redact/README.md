# redact

String redaction for log lines. Each helper keeps just enough of the
original visible to remain useful in debugging.

```go
redact.Mask("hello", 1, 1)     // "h***o"
redact.Email("alice@example.com") // "a****@example.com"
redact.Phone("13800138000")       // "138****8000"
redact.Token("abcd12345678wxyz")  // "abcd****wxyz"
```

For structured logging:

```go
zap.L().Info("auth",
    redact.ZapString("token", token, 4, 4),
)

slog.Info("auth",
    redact.SlogString("token", token, 4, 4),
)
```
