package caseconv_test

import (
	"fmt"

	"github.com/liguangsheng/goost/caseconv"
)

func ExampleCamelSplit() {
	fmt.Println(caseconv.CamelSplit("HelloHTTPWorld"))
	// Output: [Hello HTTP World]
}

func ExampleUpperCamelJoin() {
	fmt.Println(caseconv.UpperCamelJoin([]string{"i", "love", "you"}))
	// Output: ILoveYou
}

func ExampleLowerSnakeJoin() {
	fmt.Println(caseconv.LowerSnakeJoin([]string{"Hello", "World"}))
	// Output: hello_world
}
