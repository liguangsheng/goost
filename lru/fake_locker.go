package lru

type noopLocker struct{}

func (noopLocker) Lock()   {}
func (noopLocker) Unlock() {}
