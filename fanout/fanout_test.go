package fanout

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AllSubscribersReceive(t *testing.T) {
	// Buffer >= number of published messages so no Publish can drop
	// regardless of how slow the readers schedule.
	b := New[int]().Buffer(32).Build()

	const subs = 5
	var wg sync.WaitGroup
	got := make([][]int, subs)
	for i := range subs {
		s := b.Subscribe()
		wg.Add(1)
		go func(i int, s *Sub[int]) {
			defer wg.Done()
			for v := range s.C() {
				got[i] = append(got[i], v)
			}
		}(i, s)
	}

	for i := range 10 {
		b.Publish(i)
	}
	b.Close()
	wg.Wait()

	for i, g := range got {
		assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, g,
			"subscriber %d should have received all messages", i)
	}
}

func Test_SubscribeAfterPublishMissesPast(t *testing.T) {
	b := New[int]().Build()
	b.Publish(1)
	b.Publish(2)

	s := b.Subscribe()
	defer s.Close()

	go b.Publish(3)
	select {
	case v := <-s.C():
		assert.Equal(t, 3, v)
	case <-time.After(time.Second):
		t.Fatal("did not receive post-Subscribe message")
	}
}

func Test_SlowSubscriberDropsNotBlocks(t *testing.T) {
	// Buffer of 4: small enough that an unread subscriber overflows
	// quickly. The fast subscriber is acknowledged after every publish
	// so this test does not depend on goroutine scheduling fairness.
	b := New[int]().Buffer(4).Build()
	slow := b.Subscribe() // never reads
	defer slow.Close()
	fast := b.Subscribe()
	defer fast.Close()

	var fastCount atomic.Int64
	go func() {
		for range fast.C() {
			fastCount.Add(1)
		}
	}()

	const n = 100
	var publishElapsed time.Duration
	for i := range n {
		start := time.Now()
		b.Publish(i)
		publishElapsed += time.Since(start)
		require.Eventually(t, func() bool {
			return fastCount.Load() == int64(i+1)
		}, time.Second, time.Millisecond)
	}

	assert.Less(t, publishElapsed, 100*time.Millisecond,
		"Publish must not block on slow subscriber")
	assert.Greater(t, slow.Drops(), int64(0),
		"slow subscriber should have drops")
	assert.EqualValues(t, 0, fast.Drops(),
		"fast subscriber should not drop while it is actively drained")
}

func Test_SubCloseRemovesFromBroadcaster(t *testing.T) {
	b := New[int]().Build()
	s1 := b.Subscribe()
	s2 := b.Subscribe()
	assert.Equal(t, 2, b.Len())

	s1.Close()
	assert.Equal(t, 1, b.Len())

	// s1.C() should be closed.
	select {
	case _, ok := <-s1.C():
		assert.False(t, ok)
	case <-time.After(time.Second):
		t.Fatal("closed sub channel should yield (zero, false) immediately")
	}

	// Publish still works for s2.
	go b.Publish(42)
	v := <-s2.C()
	assert.Equal(t, 42, v)

	s2.Close()
	assert.Equal(t, 0, b.Len())
}

func Test_SubCloseIdempotent(t *testing.T) {
	b := New[int]().Build()
	s := b.Subscribe()
	s.Close()
	assert.NotPanics(t, s.Close, "second Close must be a no-op")
}

func Test_BroadcasterCloseClosesAllSubs(t *testing.T) {
	b := New[int]().Build()
	subs := make([]*Sub[int], 3)
	for i := range subs {
		subs[i] = b.Subscribe()
	}
	b.Close()

	for i, s := range subs {
		select {
		case _, ok := <-s.C():
			assert.False(t, ok, "sub %d channel should be closed", i)
		case <-time.After(time.Second):
			t.Fatalf("sub %d channel not closed after Broadcaster.Close", i)
		}
	}

	// Publish after Close is a no-op (but increments counter).
	b.Publish(1)
	// Subscribe after Close yields a pre-closed channel.
	s := b.Subscribe()
	_, ok := <-s.C()
	assert.False(t, ok)
}

func Test_BroadcasterCloseIdempotent(t *testing.T) {
	b := New[int]().Build()
	b.Subscribe()
	b.Close()
	assert.NotPanics(t, b.Close)
}

func Test_Stats(t *testing.T) {
	b := New[int]().Buffer(1).Build()
	s := b.Subscribe()
	defer s.Close()

	b.Publish(1)
	b.Publish(2) // buffer is 1, so second drops
	b.Publish(3) // drops

	st := b.Stats()
	assert.EqualValues(t, 3, st.Publishes)
	assert.EqualValues(t, 2, st.Drops)
	assert.Equal(t, 1, st.Subscribers)
	assert.Equal(t, 1, st.Buffer)
	assert.Equal(t, 1, st.Queued)
	assert.False(t, st.Closed)
	assert.EqualValues(t, 2, s.Drops())
}

func Test_StatsReportsClosedState(t *testing.T) {
	b := New[int]().Buffer(3).Build()
	s1 := b.Subscribe()
	s2 := b.Subscribe()

	b.Publish(1)
	st := b.Stats()
	assert.Equal(t, 2, st.Subscribers)
	assert.Equal(t, 3, st.Buffer)
	assert.Equal(t, 2, st.Queued)
	assert.False(t, st.Closed)

	s1.Close()
	st = b.Stats()
	assert.Equal(t, 1, st.Subscribers)
	assert.Equal(t, 1, st.Queued)
	assert.False(t, st.Closed)

	s2.Close()
	b.Close()
	st = b.Stats()
	assert.Equal(t, 0, st.Subscribers)
	assert.Equal(t, 0, st.Queued)
	assert.True(t, st.Closed)
}

func Test_StatsKeepsDropsAfterSubscriberClose(t *testing.T) {
	b := New[int]().Buffer(1).Build()
	s := b.Subscribe()

	b.Publish(1)
	b.Publish(2)
	b.Publish(3)
	assert.EqualValues(t, 2, s.Drops())
	assert.EqualValues(t, 2, b.Stats().Drops)

	s.Close()
	st := b.Stats()
	assert.EqualValues(t, 2, st.Drops)
	assert.Equal(t, 0, st.Subscribers)
	assert.Equal(t, 0, st.Queued)
}

func Test_StatsCountsPublishCallsAfterClose(t *testing.T) {
	b := New[int]().Build()
	s := b.Subscribe()

	b.Publish(1)
	b.Close()
	b.Publish(2)

	st := b.Stats()
	assert.EqualValues(t, 2, st.Publishes)
	assert.True(t, st.Closed)
	select {
	case v, ok := <-s.C():
		assert.True(t, ok)
		assert.Equal(t, 1, v)
	default:
		t.Fatal("first publish should remain queued after close")
	}
	select {
	case _, ok := <-s.C():
		assert.False(t, ok)
	default:
		t.Fatal("subscriber channel should be closed after queued value")
	}
}

func Test_ConcurrentPublishSubscribeClose(t *testing.T) {
	b := New[int]().Buffer(4).Build()
	const subs = 10
	const publishers = 4
	const perPub = 200

	var wg sync.WaitGroup

	// Subscribers: keep draining.
	subList := make([]*Sub[int], subs)
	for i := range subs {
		s := b.Subscribe()
		subList[i] = s
		wg.Add(1)
		go func(s *Sub[int]) {
			defer wg.Done()
			for range s.C() {
				// drain
			}
		}(s)
	}

	// Publishers.
	for p := range publishers {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			for i := range perPub {
				b.Publish(p*perPub + i)
			}
		}(p)
	}

	// Let publishers run, then close all subs.
	time.Sleep(20 * time.Millisecond)
	for _, s := range subList {
		s.Close()
	}
	b.Close()
	wg.Wait()
}
