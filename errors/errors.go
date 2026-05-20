// Package errors augments the standard library errors with stack traces
// and lightweight wrap helpers.
//
// New / Errorf / WithStack capture the call site. Wrap adds context with
// an additional capture point. The result is still compatible with
// errors.Is / errors.As / errors.Unwrap from the standard library and
// formats nicely with %+v to show frames.
package errors

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
)

const maxDepth = 32

// frames captures a stack trace at construction time.
type frames []uintptr

func captureFrames(skip int) frames {
	pcs := make([]uintptr, maxDepth)
	n := runtime.Callers(skip+2, pcs)
	return pcs[:n]
}

// Format prints frames in the same shape as github.com/pkg/errors:
//
//	function-name
//	    file:line
func (f frames) format(w io.Writer) {
	if len(f) == 0 {
		return
	}
	fr := runtime.CallersFrames(f)
	for {
		frame, more := fr.Next()
		_, _ = fmt.Fprintf(w, "\n%s\n\t%s:%d", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
}

type stackError struct {
	msg    string
	cause  error
	frames frames
}

func (e *stackError) Error() string {
	if e.cause == nil {
		return e.msg
	}
	if e.msg == "" {
		return e.cause.Error()
	}
	return e.msg + ": " + e.cause.Error()
}

func (e *stackError) Unwrap() error { return e.cause }

// Format implements fmt.Formatter. %v / %s = Error(); %+v also prints the
// stack trace.
func (e *stackError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, e.Error())
			e.frames.format(s)
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.Error())
	}
}

// New returns an error with msg and a stack captured at the call site.
func New(msg string) error {
	return &stackError{msg: msg, frames: captureFrames(1)}
}

// Errorf works like fmt.Errorf but also captures a stack trace. If the
// format string contains %w, the wrapped error is preserved.
func Errorf(format string, args ...any) error {
	formatted := fmt.Errorf(format, args...)
	return &stackError{
		msg:    formatted.Error(),
		cause:  errors.Unwrap(formatted),
		frames: captureFrames(1),
	}
}

// WithStack attaches a stack trace to err. If err is nil, returns nil. If
// err already carries a stack from this package, the existing stack is
// preserved.
func WithStack(err error) error {
	if err == nil {
		return nil
	}
	var s *stackError
	if errors.As(err, &s) {
		return err
	}
	return &stackError{cause: err, frames: captureFrames(1)}
}

// Wrap annotates err with msg and a stack trace. Wrapping nil returns nil.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &stackError{msg: msg, cause: err, frames: captureFrames(1)}
}

// Wrapf is the formatted variant of Wrap.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &stackError{msg: fmt.Sprintf(format, args...), cause: err, frames: captureFrames(1)}
}

// StackTrace returns the captured PCs, or nil if err carries no stack.
func StackTrace(err error) []uintptr {
	var s *stackError
	if errors.As(err, &s) {
		return s.frames
	}
	return nil
}

// FormatStack renders a stack trace for err as a multi-line string. Empty
// when err has no captured stack.
func FormatStack(err error) string {
	var s *stackError
	if !errors.As(err, &s) {
		return ""
	}
	var b strings.Builder
	s.frames.format(&b)
	return b.String()
}
