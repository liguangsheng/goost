package lru_test

import (
	"fmt"

	"github.com/liguangsheng/goost/lru"
)

func ExampleCache() {
	c := lru.New[string, int]().Cap(2).Build()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3) // evicts "a"

	_, hasA := c.Get("a")
	v, hasC := c.Get("c")
	fmt.Println(hasA, hasC, v)
	// Output: false true 3
}

func ExampleShardedCache() {
	c := lru.New[string, int]().Cap(64).Shards(4, lru.StringHash[string]).BuildSharded()
	c.Set("hello", 1)
	v, _ := c.Get("hello")
	fmt.Println(v)
	// Output: 1
}
