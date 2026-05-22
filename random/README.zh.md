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

令牌、盐值等安全敏感字符串请使用 `SecureString`，它从 `crypto/rand` 取随机数：

```go
token := random.SecureString(32, random.HumanAlphanumeric)
```
