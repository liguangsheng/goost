package singleflight

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_DeduplicatesConcurrentCalls(t *testing.T) {
	g := NewString[int]()
	var calls atomic.Int64

	const n = 50
	var wg sync.WaitGroup
	results := make([]int, n)
	shared := make([]bool, n)
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v, _, s := g.Do("k", func() (int, error) {
				calls.Add(1)
				time.Sleep(20 * time.Millisecond)
				return 42, nil
			})
			results[i] = v
			shared[i] = s
		}(i)
	}
	wg.Wait()

	assert.EqualValues(t, 1, calls.Load())
	for _, r := range results {
		assert.Equal(t, 42, r)
	}
	any := false
	for _, s := range shared {
		if s {
			any = true
			break
		}
	}
	assert.True(t, any, "at least one caller should see shared=true")
}

func Test_PropagatesError(t *testing.T) {
	g := NewString[int]()
	want := errors.New("oh no")
	_, err, _ := g.Do("k", func() (int, error) { return 0, want })
	assert.ErrorIs(t, err, want)
}

func Test_GenericKey(t *testing.T) {
	type id struct{ v int }
	g := New[id, string](func(k id) string { return string(rune(k.v)) })
	v, _, _ := g.Do(id{1}, func() (string, error) { return "ok", nil })
	assert.Equal(t, "ok", v)
}

func Test_Forget(t *testing.T) {
	g := NewString[int]()
	hold := make(chan struct{})
	var calls atomic.Int64

	go func() {
		_, _, _ = g.Do("k", func() (int, error) {
			calls.Add(1)
			<-hold
			return 1, nil
		})
	}()
	// give the first call time to register
	time.Sleep(10 * time.Millisecond)
	g.Forget("k")

	_, _, _ = g.Do("k", func() (int, error) {
		calls.Add(1)
		return 2, nil
	})
	close(hold)
	time.Sleep(10 * time.Millisecond)

	assert.EqualValues(t, 2, calls.Load())
}
