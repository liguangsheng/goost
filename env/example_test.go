package env_test

import (
	"fmt"
	"time"

	"github.com/liguangsheng/goost/env"
)

func ExampleLoadFromMap() {
	type Config struct {
		Addr    string        `env:"HTTP_ADDR,default=:8080"`
		Token   string        `env:"API_TOKEN,required"`
		Tags    []string      `env:"TAGS"`
		Timeout time.Duration `env:"TIMEOUT,default=5s"`
	}

	var cfg Config
	err := env.LoadFromMap(&cfg, map[string]string{
		"API_TOKEN": "secret",
		"TAGS":      "api, worker",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Addr)
	fmt.Println(cfg.Token)
	fmt.Println(cfg.Tags)
	fmt.Println(cfg.Timeout)

	// Output:
	// :8080
	// secret
	// [api worker]
	// 5s
}
