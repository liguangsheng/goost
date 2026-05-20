package errors

import (
	stderrors "errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Recover_StringPanic(t *testing.T) {
	err := callRecover(func() { panic("boom") })
	require.Error(t, err)

	var pe *PanicError
	require.True(t, stderrors.As(err, &pe))
	assert.Equal(t, "boom", pe.Value)
	assert.Contains(t, err.Error(), "boom")
}

func Test_Recover_ErrorPanic(t *testing.T) {
	sentinel := New("ouch")
	err := callRecover(func() { panic(sentinel) })
	require.Error(t, err)

	assert.True(t, stderrors.Is(err, sentinel),
		"errors.Is should reach the original error via Unwrap")

	var pe *PanicError
	require.True(t, stderrors.As(err, &pe))
	_, isErr := pe.Value.(error)
	assert.True(t, isErr)
}

func Test_Recover_NoPanicIsNoop(t *testing.T) {
	err := callRecover(func() {})
	assert.NoError(t, err)
}

func Test_Recover_PreservesExistingError(t *testing.T) {
	existing := stderrors.New("existing")
	err := func() (err error) {
		defer Recover(&err)
		err = existing
		panic("late panic")
	}()
	require.Error(t, err)
	assert.True(t, stderrors.Is(err, existing),
		"existing error must remain in the chain")
	var pe *PanicError
	assert.True(t, stderrors.As(err, &pe),
		"PanicError must also be in the chain")
}

func Test_Recover_NilErrpSwallowsPanic(t *testing.T) {
	// Calling Recover(nil) is a misuse, but it should not panic itself.
	assert.NotPanics(t, func() {
		func() {
			defer Recover(nil)
			panic("swallowed")
		}()
	})
}

func Test_Recover_StackCaptured(t *testing.T) {
	err := callRecover(func() { panic("with-stack") })
	var pe *PanicError
	require.True(t, stderrors.As(err, &pe))
	assert.NotEmpty(t, pe.Stack)
	// The stack should mention this test function name.
	assert.True(t, strings.Contains(string(pe.Stack), "Test_Recover_StackCaptured") ||
		strings.Contains(string(pe.Stack), "goroutine"),
		"stack should contain goroutine header")
}

func Test_Recover_PlusVPrintsStack(t *testing.T) {
	err := callRecover(func() { panic("fmt-test") })
	var pe *PanicError
	require.True(t, stderrors.As(err, &pe))
	plain := fmt.Sprintf("%v", pe)
	withStack := fmt.Sprintf("%+v", pe)
	assert.Contains(t, plain, "fmt-test")
	assert.Greater(t, len(withStack), len(plain), "%+v should be longer than %v")
}

// callRecover runs fn under a deferred Recover and returns the result.
func callRecover(fn func()) (err error) {
	defer Recover(&err)
	fn()
	return nil
}
