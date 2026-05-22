package shutdown_test

import (
	"fmt"

	"github.com/liguangsheng/goost/shutdown"
)

func ExampleNewManager() {
	m := shutdown.NewManager()
	m.SetLogger(nil)

	var closed []string
	m.Add(func() { closed = append(closed, "server") }, shutdown.WithName("server"))
	m.Add(func() { closed = append(closed, "database") }, shutdown.WithName("database"))

	m.Cleanup()
	m.Cleanup()

	fmt.Println(closed)

	// Output:
	// [server database]
}
