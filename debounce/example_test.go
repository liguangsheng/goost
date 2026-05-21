package debounce_test

import (
	"fmt"
	"time"

	"github.com/liguangsheng/goost/clock"
	"github.com/liguangsheng/goost/debounce"
)

func Example() {
	m := clock.NewMock(time.Unix(0, 0))
	d := debounce.New[string](100 * time.Millisecond).WithClock(m)
	defer d.Stop()

	d.Trigger("first")
	d.Trigger("latest")
	m.Advance(100 * time.Millisecond)

	fmt.Println(<-d.C())

	// Output:
	// latest
}
