package backoff

import (
	"errors"
	"testing"
)

func TestPermanentErrorIsCompatible(t *testing.T) {
	inner := errors.New("inner")
	pe := &PermanentError{Err: inner}

	if !errors.Is(pe, inner) {
		t.Error("PermanentError should unwrap to inner error via errors.Is")
	}

	var extracted *PermanentError
	if !errors.As(pe, &extracted) {
		t.Error("should be able to extract *PermanentError via errors.As")
	}
	if extracted.Err.Error() != "inner" {
		t.Errorf("expected inner, got %v", extracted.Err)
	}
}
