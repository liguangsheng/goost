package circuitbreaker

import (
	"errors"
	"testing"
)

func TestErrOpenIsCompatible(t *testing.T) {
	if !errors.Is(ErrOpen, ErrOpen) {
		t.Error("ErrOpen should match itself via errors.Is")
	}
}
