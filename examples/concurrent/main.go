// concurrent demonstrates taskgroup + backoff: N tasks run in parallel
// with a concurrency cap; each retries transient failures with
// exponential backoff before giving up.
//
// Run: go run ./examples/concurrent
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/liguangsheng/goost/backoff"
	"github.com/liguangsheng/goost/taskgroup"
)

var errFlaky = errors.New("flaky")

func fetch(ctx context.Context, id int) (string, error) {
	if rand.IntN(3) == 0 {
		return "", errFlaky
	}
	return fmt.Sprintf("item-%d", id), nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	g := taskgroup.New(ctx).WithLimit(4)
	for i := range 20 {
		g.Go(func(ctx context.Context) error {
			b := &backoff.Backoff{
				Initial: 10 * time.Millisecond,
				Max:     200 * time.Millisecond,
				Factor:  2,
				Jitter:  0.2,
			}
			return backoff.Retry(ctx, b, 5, func(ctx context.Context) error {
				v, err := fetch(ctx, i)
				if err != nil {
					return err
				}
				fmt.Println(v)
				return nil
			})
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Println("group error:", err)
	}
}
