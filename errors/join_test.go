package errors

import (
	stderrors "errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_JoinNilsReturnNil(t *testing.T) {
	assert.Nil(t, Join(nil, nil))
	assert.Nil(t, Join())
}

func Test_JoinIsCompatible(t *testing.T) {
	a := stderrors.New("a")
	b := stderrors.New("b")
	joined := Join(a, b)
	assert.True(t, stderrors.Is(joined, a))
	assert.True(t, stderrors.Is(joined, b))
}

func Test_JoinCarriesStack(t *testing.T) {
	a := stderrors.New("a")
	b := stderrors.New("b")
	joined := Join(a, b)
	assert.NotEmpty(t, StackTrace(joined))
}

func Test_JoinFormatPlusV(t *testing.T) {
	a := New("first")
	b := New("second")
	out := JoinFormatPlusV(Join(a, b))
	assert.True(t, strings.Contains(out, "first"))
	assert.True(t, strings.Contains(out, "second"))
	assert.True(t, strings.Contains(out, "---"), "expected separator between joined errors")
	assert.True(t, strings.Contains(out, "join_test.go"), "expected stack frames in output")
}
