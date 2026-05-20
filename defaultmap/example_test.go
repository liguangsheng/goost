package defaultmap_test

import (
	"fmt"

	"github.com/liguangsheng/goost/defaultmap"
)

func ExampleMap() {
	counts := defaultmap.Make(func(k string) int { return 0 })
	counts.Set("foo", counts.Get("foo")+1)
	counts.Set("foo", counts.Get("foo")+1)
	fmt.Println(counts.Get("foo"), counts.Get("bar"))
	// Output: 2 0
}
