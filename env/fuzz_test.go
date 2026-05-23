package env

import (
	"testing"
	"time"
)

func FuzzLoadFromMap(f *testing.F) {
	f.Add("KEY", "value")
	f.Add("KEY", "123")
	f.Add("KEY", "true")
	f.Add("KEY", "3.14")
	f.Add("KEY", "1h30m")
	f.Add("KEY", "a,b,c")
	f.Add("KEY", "")

	f.Fuzz(func(t *testing.T, key, val string) {
		type cfg struct {
			S string `env:"KEY"`
		}
		var c cfg
		_ = LoadFromMap(&c, map[string]string{key: val})
	})
}

func FuzzLoadFromMapInt(f *testing.F) {
	f.Add("N", "0")
	f.Add("N", "42")
	f.Add("N", "-1")
	f.Add("N", "999999999")
	f.Add("N", "notanumber")

	f.Fuzz(func(t *testing.T, key, val string) {
		type cfg struct {
			N int `env:"N"`
		}
		var c cfg
		_ = LoadFromMap(&c, map[string]string{key: val})
	})
}

func FuzzLoadFromMapBool(f *testing.F) {
	f.Add("B", "true")
	f.Add("B", "false")
	f.Add("B", "1")
	f.Add("B", "0")
	f.Add("B", "maybe")

	f.Fuzz(func(t *testing.T, key, val string) {
		type cfg struct {
			B bool `env:"B"`
		}
		var c cfg
		_ = LoadFromMap(&c, map[string]string{key: val})
	})
}

func FuzzLoadFromMapDuration(f *testing.F) {
	f.Add("D", "1s")
	f.Add("D", "1h30m")
	f.Add("D", "0")
	f.Add("D", "-5s")
	f.Add("D", "notaduration")

	f.Fuzz(func(t *testing.T, key, val string) {
		type inner struct {
			D time.Duration `env:"D"`
		}
		var c inner
		_ = LoadFromMap(&c, map[string]string{key: val})
	})
}
