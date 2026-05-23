package batcher

import (
	"errors"
	"testing"
)

func TestErrNotFoundIsCompatible(t *testing.T) {
	if !errors.Is(ErrNotFound, ErrNotFound) {
		t.Error("ErrNotFound should match itself via errors.Is")
	}
}
