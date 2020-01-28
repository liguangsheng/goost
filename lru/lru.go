package lru

import (
	"sync"
	"time"
)

type (
	LRU      = *lru
	key      = string
	value    = interface{}
	HookFunc func(k key, v value)

	entry struct {
		key       key
		value     value
		expiredAt int64
	}

	lru struct {
		access     map[key]*element
		ll         *lis
		maxEntries int
		evictHook  HookFunc
		freeList   *lis
		lock       sync.Locker
	}
)

func newLRU(maxEntries int, lock sync.Locker, hook ...HookFunc) *lru {
	e := &lru{
		maxEntries: maxEntries,
		access:     make(map[key]*element, maxEntries),
		ll:         newList(),
		freeList:   newList(),
		lock:       lock,
	}

	if len(hook) > 0 {
		e.evictHook = hook[0]
	}
	return e
}

func (e *lru) getFreeElement() *element {
	if e.freeList.Len() > 0 {
		return e.freeList.PopBack()
	}
	return &element{Value: &entry{}}
}

func (e *lru) putFreeElement(ele *element) {
	if e.ll.Len()+e.freeList.Len() <= e.maxEntries {
		e.freeList.PushBack(ele)
	}
}

func (e *lru) set(key key, value value, expiration int64) {
	if ele, ok := e.access[key]; ok {
		ele.Value.(*entry).value = value
		e.ll.MoveToFront(ele)
		return
	}

	ele := e.getFreeElement()
	ent := ele.Value.(*entry)
	ent.key = key
	ent.value = value
	ent.expiredAt = expiration
	e.access[key] = e.ll.PushFront(ele)

	if e.maxEntries != 0 && e.ll.Len() > e.maxEntries {
		e.removeOldest()
	}
}

func (e *lru) get(key key) *entry {
	ele, ok := e.access[key]
	if !ok {
		return nil
	}

	ent := ele.Value.(*entry)
	if ent.expiredAt > 0 && ent.expiredAt < time.Now().Unix() {
		return nil
	}

	e.ll.MoveToFront(ele)
	return ent
}

func (e *lru) remove(k key) {
	if ele, hit := e.access[k]; hit {
		e.removeElement(ele)
	}
}

func (e *lru) Clear() {
	e.ll = newList()
	e.access = make(map[key]*element, e.maxEntries)
}

func (e *lru) size() int {
	return e.ll.Len()
}

func (e *lru) removeOldest() {
	ele := e.ll.Back()
	if ele != nil {
		if e.evictHook != nil {
			ent := ele.Value.(*entry)
			e.evictHook(ent.key, ent.value)
		}
		e.removeElement(ele)
	}
}

func (e *lru) removeElement(ele *element) {
	ent := ele.Value.(*entry)
	key := ent.key
	e.ll.Remove(ele)
	e.putFreeElement(ele)
	delete(e.access, key)
}

func (e *lru) Set(key key, value value) {
	e.lock.Lock()
	e.set(key, value, -1)
	e.lock.Unlock()
}

func (e *lru) SetWithExpire(key key, value value, expiredAt time.Time) {
	e.lock.Lock()
	e.set(key, value, expiredAt.Unix())
	e.lock.Unlock()
}

func (e *lru) SetWithDuration(key key, value value, duration time.Duration) {
	e.lock.Lock()
	e.set(key, value, time.Now().Add(duration).Unix())
	e.lock.Unlock()
}

func (e *lru) Get(key key) (value, bool) {
	e.lock.Lock()
	ent := e.get(key)
	if ent == nil {
		e.lock.Unlock()
		return nil, false
	}
	e.lock.Unlock()
	return ent.value, true
}

func (e *lru) Remove(key key) {
	e.lock.Lock()
	e.remove(key)
	e.lock.Unlock()
}

func (e *lru) Size() int {
	return e.size()
}
