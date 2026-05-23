// concurrent demonstrates taskgroup + backoff: N tasks run in parallel
// with a concurrency cap; each retries transient failures with
// exponential backoff before giving up.
//
// Run from examples/: go run ./concurrent
package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/liguangsheng/goost/backoff"
	"github.com/liguangsheng/goost/taskgroup"
)

var errFlaky = errors.New("flaky")

func fetch(ctx context.Context, id, attempt int) (string, error) {
	if id%3 == 0 && attempt == 1 {
		return "", errFlaky
	}
	return fmt.Sprintf("item-%d", id), nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	g := taskgroup.New(ctx).WithLimit(4)
	var mu sync.Mutex
	var items []string
	for i := range 20 {
		g.Go(func(ctx context.Context) error {
			b := &backoff.Backoff{
				Initial: 10 * time.Millisecond,
				Max:     200 * time.Millisecond,
				Factor:  2,
				Jitter:  0.2,
			}
			attempt := 0
			return backoff.Retry(ctx, b, 5, func(ctx context.Context) error {
				attempt++
				v, err := fetch(ctx, i, attempt)
				if err != nil {
					return err
				}
				mu.Lock()
				items = append(items, v)
				mu.Unlock()
				return nil
			})
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Println("group error:", err)
		return
	}
	sort.Strings(items)
	fmt.Printf("processed=%d first=%s last=%s\n", len(items), items[0], items[len(items)-1])
}
