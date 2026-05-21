# random

并发安全的随机字符串生成器。默认源使用 `math/rand/v2`。

```go
random.String(16, random.HumanAlphanumeric)
// "kJ3pq8rNb2tFm7sH"
```

内置字符集：`Uppercase`、`Lowercase`、`Alphabetic`、`Numeric`、
`Alphanumeric`、`HumanAlphanumeric`、`Symbols`、`Hex`。

如果需要可复现的序列，可以用自己的随机源构造 `Sequence`：

```go
src := rand.New(rand.NewPCG(1, 2))
s := random.NewSequence(src.Uint64)
s.Next(8, random.Hex)
```

> 该生成器适合非安全用途。令牌等安全场景请使用 `crypto/rand`。
