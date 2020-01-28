package lru

type element struct {
	next, prev *element
	list       *lis
	Value      interface{}
}

func (e *element) Next() *element {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

func (e *element) Prev() *element {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

type lis struct {
	root element
	len  int
}

func (l *lis) Init() *lis {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

func newList() *lis { return new(lis).Init() }

func (l *lis) Len() int { return l.len }

func (l *lis) Front() *element {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

func (l *lis) Back() *element {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

func (l *lis) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

func (l *lis) insert(e, at *element) *element {
	n := at.next
	at.next = e
	e.prev = at
	e.next = n
	n.prev = e
	e.list = l
	l.len++
	return e
}

func (l *lis) remove(e *element) *element {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil
	e.prev = nil
	e.list = nil
	l.len--
	return e
}

func (l *lis) Remove(e *element) {
	if e.list == l {
		e = l.remove(e)
	}
}

func (l *lis) PushFront(e *element) *element {
	l.lazyInit()
	return l.insert(e, &l.root)
}

func (l *lis) PushBack(e *element) *element {
	l.lazyInit()
	return l.insert(e, l.root.prev)
}

func (l *lis) MoveToFront(e *element) {
	if e.list != l || l.root.next == e {
		return
	}
	l.insert(l.remove(e), &l.root)
}

func (l *lis) PopBack() *element {
	return l.remove(l.Back())
}
