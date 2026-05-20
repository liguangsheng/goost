# errors

A thin layer over the standard library `errors` that captures a stack
trace at the wrap site. Fully compatible with `errors.Is` / `errors.As` /
`errors.Unwrap` and `fmt.Errorf("...: %w", err)`.

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

// Caller:
if err := read(); err != nil {
    log.Printf("%+v", err) // prints message + stack frames
}
```

| Helper | Behavior |
| --- | --- |
| `New(msg)` | Like `errors.New` plus a stack at the call site. |
| `Errorf(fmt, args...)` | Like `fmt.Errorf`; `%w` preserved; stack captured. |
| `WithStack(err)` | Attach a stack; no-op if one is already attached. |
| `Wrap(err, msg)` | Annotate + new stack. Nil in → nil out. |
| `Wrapf(err, fmt, args...)` | Formatted variant of `Wrap`. |
| `StackTrace(err)` | Returns the captured PCs (or nil). |
| `FormatStack(err)` | Renders the stack as a multi-line string. |
