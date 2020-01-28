package lru

type fakeLocker struct{}

func (l *fakeLocker) Lock() {}

func (l *fakeLocker) Unlock() {}
