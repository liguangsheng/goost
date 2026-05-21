package fanout_test

import (
	"fmt"

	"github.com/liguangsheng/goost/fanout"
)

func Example() {
	b := fanout.New[string]().Buffer(2).Build()
	sub := b.Subscribe()
	defer sub.Close()

	b.Publish("created")
	b.Publish("updated")

	fmt.Println(<-sub.C())
	fmt.Println(<-sub.C())

	// Output:
	// created
	// updated
}
