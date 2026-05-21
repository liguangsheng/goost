# env

使用 `env` 标签把环境变量加载到结构体中。

```go
type Config struct {
    Addr    string        `env:"HTTP_ADDR,default=:8080"`
    Token   string        `env:"API_TOKEN,required"`
    Debug   bool          `env:"DEBUG"`
    Tags    []string      `env:"TAGS"`        // 逗号分隔
    Timeout time.Duration `env:"TIMEOUT,default=5s"`
}

var cfg Config
if err := env.Load(&cfg); err != nil {
    log.Fatal(err)
}
```

### 标签选项

- `NAME`：环境变量名
- `default=VAL`：环境变量未设置或为空时使用的值
- `required`：环境变量未设置时返回错误

### 支持的类型

`string`、`bool`、所有整数类型、`time.Duration`、`float32/64`、
`[]string`，以及这些类型的指针（变量缺失时指针保持 nil）。

### 测试

使用 `env.LoadFromMap(&cfg, map[string]string{...})` 注入值，
无需触碰 `os.Setenv`。
