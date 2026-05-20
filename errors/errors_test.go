package errors

import (
	stderrors "errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewHasStack(t *testing.T) {
	err := New("boom")
	assert.Equal(t, "boom", err.Error())
	assert.NotEmpty(t, StackTrace(err))
	assert.Contains(t, fmt.Sprintf("%+v", err), "Test_NewHasStack")
}

func Test_WrapPreservesCause(t *testing.T) {
	inner := New("inner")
	outer := Wrap(inner, "outer")
	assert.Equal(t, "outer: inner", outer.Error())
	assert.True(t, stderrors.Is(outer, inner))
}

func Test_WrapNil(t *testing.T) {
	assert.Nil(t, Wrap(nil, "msg"))
	assert.Nil(t, Wrapf(nil, "msg %d", 1))
	assert.Nil(t, WithStack(nil))
}

func Test_ErrorfPreservesW(t *testing.T) {
	target := stderrors.New("sentinel")
	err := Errorf("ctx: %w", target)
	assert.True(t, stderrors.Is(err, target))
	assert.Contains(t, err.Error(), "sentinel")
}

func Test_WithStackIsIdempotent(t *testing.T) {
	a := New("x")
	b := WithStack(a)
	assert.Same(t, a, b)
}

func Test_FormatPlusV(t *testing.T) {
	err := Wrap(New("inner"), "outer")
	out := fmt.Sprintf("%+v", err)
	// Outer error message present, plus at least the wrap call site.
	assert.True(t, strings.HasPrefix(out, "outer: inner"))
	assert.Contains(t, out, "errors_test.go")
}

func Test_StackTraceForeignErrorIsNil(t *testing.T) {
	assert.Nil(t, StackTrace(stderrors.New("plain")))
	assert.Equal(t, "", FormatStack(stderrors.New("plain")))
}

type myErr struct{ s string }

func (m *myErr) Error() string { return m.s }

func Test_AsCompatibility(t *testing.T) {
	inner := &myErr{"deep"}
	wrapped := Wrap(inner, "context")

	var target *myErr
	assert.True(t, stderrors.As(wrapped, &target))
	assert.Equal(t, "deep", target.s)
}
