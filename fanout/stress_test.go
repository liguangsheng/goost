package fanout

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_StressSlowSubscribersDoNotBlockPublishers(t *testing.T) {
	b := New[int]().Buffer(4).Build()
	defer b.Close()

	for range 16 {
		b.Subscribe() // deliberately never drained
	}
	for range 4 {
		sub := b.Subscribe()
		go func() {
			for range sub.C() {
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		var wg sync.WaitGroup
		for p := range 8 {
			wg.Add(1)
			go func(p int) {
				defer wg.Done()
				for i := range 500 {
					b.Publish(p*500 + i)
				}
			}(p)
		}
		wg.Wait()
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Publish blocked behind slow subscribers")
	}
	assert.Greater(t, b.Stats().Drops, int64(0))
}
