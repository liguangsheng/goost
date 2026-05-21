package errors_test

import (
	stderrors "errors"
	"fmt"

	goosterrors "github.com/liguangsheng/goost/errors"
)

func ExampleRecover() {
	err := func() (err error) {
		defer goosterrors.Recover(&err)
		panic("bad input")
	}()

	var pe *goosterrors.PanicError
	fmt.Println(stderrors.As(err, &pe))
	fmt.Println(pe.Value)

	// Output:
	// true
	// bad input
}
