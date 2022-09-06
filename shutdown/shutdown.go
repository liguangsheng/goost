package shutdown

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	cleanupOnce sync.Once
	fns         []func()
	m           sync.Mutex
	once        sync.Once
	shutdownC   chan os.Signal
)

func C() {
	once.Do(func() {
		shutdownC = make(chan os.Signal, 1)
		signal.Notify(shutdownC, syscall.SIGTERM, syscall.SIGINT)
	})
	<-shutdownC
	log.Println("shutdown: receive a interrupt signal, exit")
	Cleanup()
	os.Exit(0)
}

func Add(f func()) {
	m.Lock()
	fns = append(fns, f)
	m.Unlock()
}

func Cleanup() {
	cleanupOnce.Do(func() {
		log.Printf("shutdown: performing %d cleanups\n", len(fns))
		for _, f := range fns {
			f()
		}
		log.Println("shutdown: all Cleanup done.")
	})
}
