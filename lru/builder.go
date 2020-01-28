package lru

import "sync"

type builder struct {
	cap       int
	locker    sync.Locker
	evictHook HookFunc
}

func New() *builder {
	return &builder{
		cap:    10000,
		locker: &sync.Mutex{},
	}
}

func (b *builder) Cap(i int) *builder {
	b.cap = i
	return b
}

func (b *builder) Safe(safe bool) *builder {
	if safe {
		b.locker = &sync.Mutex{}
	} else {
		b.locker = &fakeLocker{}
	}
	return b
}

func (b *builder) Evict(fn HookFunc) *builder {
	b.evictHook = fn
	return b
}

func (b *builder) Build() LRU {
	return newLRU(b.cap, b.locker, b.evictHook)
}
