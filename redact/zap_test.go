package redact

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func Test_ZapStringMasks(t *testing.T) {
	f := ZapString("token", "abcd12345678wxyz", 4, 4)
	assert.Equal(t, "token", f.Key)
	assert.Equal(t, "abcd********wxyz", f.String)
	assert.Equal(t, zapcore.StringType, f.Type)
}

func Test_SlogStringMasks(t *testing.T) {
	a := SlogString("token", "abcd12345678wxyz", 4, 4)
	assert.Equal(t, "token", a.Key)
	assert.Equal(t, "abcd********wxyz", a.Value.String())
}
