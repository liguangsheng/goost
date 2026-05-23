package ratelimit

import (
	"errors"
	"testing"
)

func TestErrLimitExceededIsCompatible(t *testing.T) {
	if !errors.Is(ErrLimitExceeded, ErrLimitExceeded) {
		t.Error("ErrLimitExceeded should match itself via errors.Is")
	}
}
