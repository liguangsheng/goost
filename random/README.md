# random

Concurrency-safe random string generator. Uses `math/rand/v2` for the
default source.

```go
random.String(16, random.HumanAlphanumeric)
// "kJ3pq8rNb2tFm7sH"
```

Built-in charsets: `Uppercase`, `Lowercase`, `Alphabetic`, `Numeric`,
`Alphanumeric`, `HumanAlphanumeric`, `Symbols`, `Hex`.

For a reproducible stream, build a `Sequence` with your own source:

```go
src := rand.New(rand.NewPCG(1, 2))
s := random.NewSequence(src.Uint64)
s.Next(8, random.Hex)
```

For tokens, salts, and other security-sensitive strings, use `SecureString`,
which draws from `crypto/rand`:

```go
token := random.SecureString(32, random.HumanAlphanumeric)
```
