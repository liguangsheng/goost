package shutdown

import (
	"github.com/golang/glog"
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
	glog.Info("shutdown: receive a interrupt signal, exit")
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
		glog.Infof("shutdown: performing %d cleanups", len(fns))
		for _, f := range fns {
			f()
		}
		glog.Infof("shutdown: all Cleanup done.")
		glog.Flush()
	})
}
