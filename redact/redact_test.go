package redact

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Mask(t *testing.T) {
	assert.Equal(t, "h***o", Mask("hello", 1, 1))
	assert.Equal(t, "*****", Mask("hello", 5, 0))
	assert.Equal(t, "*****", Mask("hello", 10, 0))
	assert.Equal(t, "*****", Mask("hello", 0, 0))
	assert.Equal(t, "", Mask("", 1, 1))
}

func Test_Email(t *testing.T) {
	assert.Equal(t, "a****@example.com", Email("alice@example.com"))
	assert.Equal(t, "n*******", Email("noatsign"))
	assert.Equal(t, "*@example.com", Email("a@example.com"))
}

func Test_Phone(t *testing.T) {
	assert.Equal(t, "138****8000", Phone("13800138000"))
}

func Test_Token(t *testing.T) {
	assert.Equal(t, "abcd********wxyz", Token("abcd12345678wxyz"))
}
