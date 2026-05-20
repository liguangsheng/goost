package errors

import (
	stderrors "errors"
	"io"
	"strings"
)

// Join returns an error that wraps the given errors using stdlib
// errors.Join and attaches a stack at the call site. Nil errors are
// dropped. If every input is nil, Join returns nil.
func Join(errs ...error) error {
	joined := stderrors.Join(errs...)
	if joined == nil {
		return nil
	}
	return &stackError{
		msg:    joined.Error(),
		cause:  joined,
		frames: captureFrames(1),
	}
}

// JoinFormatPlusV renders err with a "+v" verb. For a joined error this
// expands each underlying error on its own line, with its own stack
// frames if available. Useful when fmt.Sprintf("%+v", err) elides nested
// stacks because the outer wrapper only carries its own.
func JoinFormatPlusV(err error) string {
	if err == nil {
		return ""
	}
	var b strings.Builder
	formatChain(&b, err)
	return b.String()
}

func formatChain(b *strings.Builder, err error) {
	type multi interface{ Unwrap() []error }

	// Look down the chain for the first multi-error; if found, expand it.
	for cur := err; cur != nil; {
		if m, ok := cur.(multi); ok {
			children := m.Unwrap()
			for i, e := range children {
				if i > 0 {
					b.WriteString("\n---\n")
				}
				formatChain(b, e)
			}
			// Append the outermost stack on its own line.
			if s := FormatStack(err); s != "" {
				b.WriteString(s)
			}
			return
		}
		unwrapped, ok := cur.(interface{ Unwrap() error })
		if !ok {
			break
		}
		cur = unwrapped.Unwrap()
	}

	_, _ = io.WriteString(b, err.Error())
	if s := FormatStack(err); s != "" {
		b.WriteString(s)
	}
}
