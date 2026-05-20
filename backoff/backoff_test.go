package backoff

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_BackoffSequence(t *testing.T) {
	b := &Backoff{Initial: 10 * time.Millisecond, Max: 80 * time.Millisecond, Factor: 2}
	want := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		40 * time.Millisecond,
		80 * time.Millisecond,
		80 * time.Millisecond,
	}
	for _, w := range want {
		assert.Equal(t, w, b.Next())
	}
}

func Test_BackoffReset(t *testing.T) {
	b := &Backoff{Initial: 10 * time.Millisecond}
	b.Next()
	b.Next()
	b.Reset()
	assert.Equal(t, 10*time.Millisecond, b.Next())
}

func Test_BackoffJitterBounded(t *testing.T) {
	b := &Backoff{Initial: 100 * time.Millisecond, Max: 100 * time.Millisecond, Factor: 2, Jitter: 0.2}
	for range 50 {
		d := b.Next()
		assert.GreaterOrEqual(t, d, 80*time.Millisecond)
		assert.LessOrEqual(t, d, 120*time.Millisecond)
	}
}

func Test_RetrySuccess(t *testing.T) {
	var calls atomic.Int64
	err := Retry(context.Background(), &Backoff{Initial: time.Millisecond}, 5, func(_ context.Context) error {
		if calls.Add(1) < 3 {
			return errors.New("transient")
		}
		return nil
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 3, calls.Load())
}

func Test_RetryMaxAttempts(t *testing.T) {
	var calls atomic.Int64
	err := Retry(context.Background(), &Backoff{Initial: time.Millisecond}, 3, func(_ context.Context) error {
		calls.Add(1)
		return errors.New("nope")
	})
	assert.EqualError(t, err, "nope")
	assert.EqualValues(t, 3, calls.Load())
}

func Test_RetryPermanent(t *testing.T) {
	want := errors.New("fatal")
	var calls atomic.Int64
	err := Retry(context.Background(), &Backoff{Initial: time.Millisecond}, 0, func(_ context.Context) error {
		calls.Add(1)
		return Permanent(want)
	})
	assert.ErrorIs(t, err, want)
	assert.EqualValues(t, 1, calls.Load())
}

func Test_RetryContextCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	var calls atomic.Int64
	err := Retry(ctx, &Backoff{Initial: 10 * time.Millisecond}, 0, func(_ context.Context) error {
		calls.Add(1)
		return errors.New("retry me")
	})
	assert.Error(t, err)
	assert.Greater(t, calls.Load(), int64(0))
}
