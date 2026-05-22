package ratelimit_test

import (
	"fmt"
	"time"

	"github.com/liguangsheng/goost/ratelimit"
)

func ExampleNewBucket() {
	now := time.Unix(0, 0)
	b := ratelimit.NewBucket(10, 1)
	b.SetClock(func() time.Time { return now })

	fmt.Println(b.Allow())
	fmt.Println(b.Allow())

	now = now.Add(100 * time.Millisecond)
	fmt.Println(b.Allow())

	// Output:
	// true
	// false
	// true
}
