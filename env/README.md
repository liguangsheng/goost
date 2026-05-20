# env

Load configuration from environment variables into a struct using `env`
tags.

```go
type Config struct {
    Addr    string        `env:"HTTP_ADDR,default=:8080"`
    Token   string        `env:"API_TOKEN,required"`
    Debug   bool          `env:"DEBUG"`
    Tags    []string      `env:"TAGS"`        // comma-separated
    Timeout time.Duration `env:"TIMEOUT,default=5s"`
}

var cfg Config
if err := env.Load(&cfg); err != nil {
    log.Fatal(err)
}
```

### Tag options

- `NAME` — name of the environment variable
- `default=VAL` — value used when the env var is unset or empty
- `required` — return an error if the env var is unset

### Supported types

`string`, `bool`, all integer kinds, `time.Duration`, `float32/64`,
`[]string`, and pointers to those (a missing var leaves the pointer nil).

### Testing

Use `env.LoadFromMap(&cfg, map[string]string{...})` to inject values
without touching `os.Setenv`.
