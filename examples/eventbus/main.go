// eventbus shows debounce + fanout cooperating: a noisy producer (think
// inotify firing several times per save) feeds a debouncer; the
// debouncer's quiet-window emit is published to a broadcaster; multiple
// subscribers (reloader, audit log, metrics) each receive the same
// event independently.
//
// Run: go run ./examples/eventbus
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/liguangsheng/goost/debounce"
	"github.com/liguangsheng/goost/fanout"
)

type ConfigEvent struct {
	Path string
	When time.Time
}

func main() {
	deb := debounce.New[ConfigEvent](80 * time.Millisecond)
	defer deb.Stop()

	bus := fanout.New[ConfigEvent]().Buffer(8).Build()
	defer bus.Close()

	// Wire debounce -> bus.
	go func() {
		for ev := range deb.C() {
			bus.Publish(ev)
		}
	}()

	// Three subscribers, each doing different work.
	var reloads, audits, metrics atomic.Int64

	subs := []*fanout.Sub[ConfigEvent]{bus.Subscribe(), bus.Subscribe(), bus.Subscribe()}
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for ev := range subs[0].C() {
			_ = ev // reloader: re-parse config (cheap here)
			reloads.Add(1)
		}
	}()
	go func() {
		defer wg.Done()
		for ev := range subs[1].C() {
			fmt.Printf("[audit] config changed: %s at %s\n", ev.Path, ev.When.Format("15:04:05.000"))
			audits.Add(1)
		}
	}()
	go func() {
		defer wg.Done()
		for range subs[2].C() {
			metrics.Add(1) // bump a counter
		}
	}()

	// Producer: simulate a save flurry (5 events in 50ms), pause, then
	// another flurry. Each flurry should debounce to one event.
	flurry := func(path string) {
		for range 5 {
			deb.Trigger(ConfigEvent{Path: path, When: time.Now()})
			time.Sleep(10 * time.Millisecond)
		}
	}
	flurry("app.yaml")
	time.Sleep(200 * time.Millisecond)
	flurry("db.yaml")
	time.Sleep(200 * time.Millisecond)

	// Drain.
	for _, s := range subs {
		s.Close()
	}
	wg.Wait()

	fmt.Printf("\nreloads=%d audits=%d metrics=%d\n",
		reloads.Load(), audits.Load(), metrics.Load())
	fmt.Println("each flurry of 5 raw events collapsed to 1 emit -> 2 total per subscriber")
}
