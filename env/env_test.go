package env

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_LoadAllSupportedTypes(t *testing.T) {
	type cfg struct {
		Name    string        `env:"NAME"`
		Port    int           `env:"PORT,default=8080"`
		Debug   bool          `env:"DEBUG"`
		Pi      float64       `env:"PI"`
		Tags    []string      `env:"TAGS"`
		Timeout time.Duration `env:"TIMEOUT,default=5s"`
		Optional *string      `env:"OPT"`
	}
	var c cfg
	err := LoadFromMap(&c, map[string]string{
		"NAME":  "goost",
		"DEBUG": "true",
		"PI":    "3.14",
		"TAGS":  "a, b ,c",
		"OPT":   "yes",
	})
	assert.NoError(t, err)
	assert.Equal(t, "goost", c.Name)
	assert.Equal(t, 8080, c.Port)
	assert.True(t, c.Debug)
	assert.Equal(t, 3.14, c.Pi)
	assert.Equal(t, []string{"a", "b", "c"}, c.Tags)
	assert.Equal(t, 5*time.Second, c.Timeout)
	if assert.NotNil(t, c.Optional) {
		assert.Equal(t, "yes", *c.Optional)
	}
}

func Test_LoadMissingRequired(t *testing.T) {
	type cfg struct {
		Token string `env:"TOKEN,required"`
	}
	var c cfg
	err := LoadFromMap(&c, map[string]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TOKEN")
}

func Test_LoadParseError(t *testing.T) {
	type cfg struct {
		N int `env:"N"`
	}
	var c cfg
	err := LoadFromMap(&c, map[string]string{"N": "not-a-number"})
	assert.Error(t, err)
}

func Test_LoadRejectsNonStruct(t *testing.T) {
	v := 0
	err := Load(&v)
	assert.Error(t, err)
}

func Test_PointerLeftNilWhenAbsent(t *testing.T) {
	type cfg struct {
		Opt *int `env:"OPT"`
	}
	var c cfg
	err := LoadFromMap(&c, map[string]string{})
	assert.NoError(t, err)
	assert.Nil(t, c.Opt)
}

func Test_LoadFromOsEnv(t *testing.T) {
	t.Setenv("GOOST_ENV_NAME", "world")
	type cfg struct {
		Name string `env:"GOOST_ENV_NAME"`
	}
	var c cfg
	assert.NoError(t, Load(&c))
	assert.Equal(t, "world", c.Name)
}
