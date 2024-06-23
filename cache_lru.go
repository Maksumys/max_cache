package max_cache

import (
	"sync"
	"sync/atomic"
)

type List[K any, V any] struct {
	Head *Node[K, V]
	Tail *Node[K, V]
}

func NewList[K any, V any]() *List[K, V] {
	return &List[K, V]{}
}

func (l *List[K, V]) InsertTail(k K, v V) *Node[K, V] {
	if l.Head == nil {
		node := NewNode(nil, nil, k, v)
		l.Head = node
		if l.Tail == nil {
			l.Tail = l.Head
		}
		return node
	}

	node := NewNode(l.Tail, nil, k, v)

	l.Tail.Next = node
	l.Tail = node

	return node
}

func (l *List[K, V]) PopTail() (v V) {
	if l.Tail == nil {
		return
	}

	v = l.Tail.V

	buf := l.Tail.Prev
	buf.Next = nil
	l.Tail = buf

	return
}

func (l *List[K, V]) InsertFront(k K, v V) {
	if l.Head == nil {
		l.Head = NewNode(nil, nil, k, v)
		if l.Tail == nil {
			l.Tail = l.Head
		}
		return
	}

	node := NewNode(nil, l.Head, k, v)
	l.Head.Prev = node
	l.Head = node
}

func (l *List[K, V]) PopFront() (k K, v V) {
	if l.Head == nil {
		return
	}

	v = l.Head.V
	k = l.Head.K
	l.Head = l.Head.Next

	if l.Head != nil {
		l.Head.Prev = nil
	}

	return
}

func (l *List[K, V]) BubbleToTail(node *Node[K, V]) {
	if node.Prev != nil {
		node.Prev.Next = node.Next
	}
	if node.Next != nil {
		node.Next.Prev = node.Prev
	}

	if l.Head == nil {
		l.Head = node
		if l.Tail == nil {
			l.Tail = l.Head
		}
		return
	}

	node.Prev = l.Tail
	l.Tail.Next = node
	l.Tail = node
}

func (l *List[K, V]) Erase(node *Node[K, V]) {
	if l.Head == l.Tail && l.Head == node {
		l.Head = nil
		l.Tail = nil
		return
	}

	if node == l.Head {
		l.Head = l.Head.Next
		l.Head.Prev = nil
		return
	}

	if node == l.Tail {
		l.Tail = l.Tail.Prev
		l.Tail.Next = nil
		return
	}

	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev

	return
}

func (l *List[K, V]) Reverse() {
	if l.Head == nil || l.Tail == nil {
		return
	}

	var prev *Node[K, V]
	iter1 := l.Head

	for iter1.Next != nil {
		iter2 := iter1.Next

		iter1.Prev = iter1.Next
		iter1.Next = prev

		prev = iter1
		iter1 = iter2
	}

	l.Tail = l.Head
	iter1.Prev = iter1.Next
	iter1.Next = prev
	l.Head = iter1
}

type Node[K any, V any] struct {
	Prev *Node[K, V]
	Next *Node[K, V]
	V    V
	K    K
}

func NewNode[K, V any](prev *Node[K, V], next *Node[K, V], k K, v V) *Node[K, V] {
	return &Node[K, V]{
		Prev: prev,
		Next: next,
		V:    v,
		K:    k,
	}
}

type Lockable interface {
	sync.Locker
}

type CacheLruLock[M Lockable, K comparable, V any] struct {
	cache   *CacheLru[K, V]
	mutex   M
	onEvict func(v V)
}

func NewCacheLruLock[M sync.Locker, K comparable, V any](capacity int, onEvict func(v V), m M) *CacheLruLock[M, K, V] {
	return &CacheLruLock[M, K, V]{
		cache:   NewCacheLru[K, V](capacity, nil),
		mutex:   m,
		onEvict: onEvict,
	}
}

func (c *CacheLruLock[M, K, V]) Set(k K, v V) {
	func() {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		c.cache.Set(k, v)
	}()

	if c.onEvict != nil {
		c.onEvict(v)
	}
}

func (c *CacheLruLock[M, K, V]) Get(k K) (v V, ok bool) {
	func() {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		v, ok = c.cache.Get(k)
	}()

	return
}

func (c *CacheLruLock[M, K, V]) Erase(k K) {
	var v V
	var ok bool

	func() {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		v, ok = c.cache.Erase(k)
	}()

	if ok && c.onEvict != nil {
		c.onEvict(v)
	}
}

type CacheLru[K comparable, V any] struct {
	list    *List[K, V]
	m       map[K]*Node[K, V]
	onEvict func(v V)

	length   int
	capacity int
}

func NewCacheLru[K comparable, V any](capacity int, onEvict func(v V)) *CacheLru[K, V] {
	return &CacheLru[K, V]{
		capacity: capacity,
		list:     NewList[K, V](),
		m:        make(map[K]*Node[K, V]),
		onEvict:  onEvict,
	}
}

func (c *CacheLru[K, V]) Set(k K, v V) {
	node, ok := c.m[k]

	if !ok {
		if c.length == c.capacity {
			delK, delV := c.list.PopFront()
			delete(c.m, delK)

			if c.onEvict != nil {
				c.onEvict(delV)
			}
		} else {
			c.length++
		}

		node = c.list.InsertTail(k, v)

		c.m[k] = node

		return
	}

	node.V = v

	c.list.BubbleToTail(node)
}

func (c *CacheLru[K, V]) Get(k K) (v V, ok bool) {
	node, ok := c.m[k]

	if !ok {
		return
	}

	c.list.BubbleToTail(node)

	return node.V, true
}

func (c *CacheLru[K, V]) Erase(k K) (v V, ok bool) {
	node, ok := c.m[k]

	if !ok {
		return
	}

	c.list.Erase(node)

	delete(c.m, k)
	c.length--

	v = node.V

	if c.onEvict != nil {
		c.onEvict(node.V)
	}

	return
}

type SpinMutex struct {
	lock atomic.Bool
}

func (m *SpinMutex) Lock() {
	a := false
	for m.lock.CompareAndSwap(a, true) {
		a = false
	}
}

func (m *SpinMutex) Unlock() {
	m.lock.Store(false)
}

func (m *SpinMutex) TryLock() bool {
	return m.lock.CompareAndSwap(false, true)
}
