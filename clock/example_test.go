package clock_test

import (
	"fmt"
	"time"

	"github.com/liguangsheng/goost/clock"
)

func ExampleMock() {
	start := time.Unix(100, 0)
	m := clock.NewMock(start)
	done := m.After(2 * time.Second)

	m.Advance(time.Second)
	fmt.Println("before:", m.Now().Sub(start))

	m.Advance(time.Second)
	fmt.Println("tick:", (<-done).Sub(start))

	// Output:
	// before: 1s
	// tick: 2s
}
