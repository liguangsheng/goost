package errors

import (
	stderrors "errors"
	"fmt"
	"io"
	"runtime/debug"
)

// PanicError wraps a value recovered from a panic. Use errors.As to
// detect it; its Unwrap returns the original value when it implemented
// error so errors.Is on the underlying error works.
type PanicError struct {
	// Value is the raw value passed to panic. Most often a string or
	// an error, but can be any type.
	Value any
	// Stack is the stack trace captured at the recovery point. Same
	// format as runtime/debug.Stack.
	Stack []byte
}

// Error returns "recovered from panic: <value>".
func (e *PanicError) Error() string {
	return fmt.Sprintf("recovered from panic: %v", e.Value)
}

// Unwrap returns the panic value if it is an error, otherwise nil.
func (e *PanicError) Unwrap() error {
	if err, ok := e.Value.(error); ok {
		return err
	}
	return nil
}

// Format implements fmt.Formatter. %+v prints the stack as well.
func (e *PanicError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = fmt.Fprintf(s, "%s\n%s", e.Error(), e.Stack)
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.Error())
	}
}

// Recover converts a recovered panic into an error written to *errp.
// Designed to be used via defer:
//
//	func foo() (err error) {
//	    defer errors.Recover(&err)
//	    mightPanic()
//	    return nil
//	}
//
// If *errp is already non-nil at recovery time, the panic error is
// joined with it (the existing error is primary; PanicError is
// appended via errors.Join). If errp is nil, the panic is silently
// swallowed — usually a mistake; prefer passing &err.
//
// If no panic is in progress, Recover does nothing.
func Recover(errp *error) {
	r := recover()
	if r == nil {
		return
	}
	pe := &PanicError{Value: r, Stack: debug.Stack()}
	if errp == nil {
		return
	}
	if *errp != nil {
		*errp = stderrors.Join(*errp, pe)
		return
	}
	*errp = pe
}
