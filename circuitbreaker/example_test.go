package circuitbreaker_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/liguangsheng/goost/circuitbreaker"
)

func ExampleBreaker() {
	now := time.Unix(0, 0)
	b := circuitbreaker.New(circuitbreaker.Config{
		FailureThreshold: 1,
		CooldownPeriod:   time.Second,
		Now:              func() time.Time { return now },
	})

	fail := errors.New("downstream failed")
	_ = b.Do(context.Background(), func(context.Context) error { return fail })
	fmt.Println(b.State())

	err := b.Do(context.Background(), func(context.Context) error { return nil })
	fmt.Println(errors.Is(err, circuitbreaker.ErrOpen))

	now = now.Add(time.Second)
	fmt.Println(b.State())

	err = b.Do(context.Background(), func(context.Context) error { return nil })
	fmt.Println(err == nil)
	fmt.Println(b.State())

	// Output:
	// open
	// true
	// half-open
	// true
	// closed
}
