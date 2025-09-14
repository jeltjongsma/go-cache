package pkg

import (
	"container/heap"
	"time"
)

type Entry[K comparable] struct {
	k         K
	expiresAt time.Time
	index     int
	seq       uint64
}

type TTLQueue[K comparable] struct {
	queue []*Entry[K]
	ttl   time.Duration
	keys  map[K]*Entry[K]
	now   func() time.Time
	seq   uint64
}

func NewTTLQueue[K comparable](defaultTTL time.Duration) *TTLQueue[K] {
	return &TTLQueue[K]{
		ttl:  defaultTTL,
		keys: make(map[K]*Entry[K]),
		now:  time.Now,
		seq:  0,
	}
}

func (t *TTLQueue[K]) Len() int {
	return len(t.queue)
}

func (t *TTLQueue[K]) Less(i, j int) bool {
	x, y := t.queue[i], t.queue[j]
	if x.expiresAt.Equal(y.expiresAt) {
		return x.seq < y.seq
	}
	return x.expiresAt.Before(y.expiresAt)
}

func (t *TTLQueue[K]) Swap(i, j int) {
	t.queue[i], t.queue[j] = t.queue[j], t.queue[i]
	t.queue[i].index = i
	t.queue[j].index = j
}

func (t *TTLQueue[K]) PushWithTTL(k K, ttl time.Duration) {
	if e, ok := t.keys[k]; ok {
		e.expiresAt = t.now().Add(ttl)
		heap.Fix(t, e.index)
		return
	}
	entry := &Entry[K]{
		k:         k,
		expiresAt: t.now().Add(ttl),
	}
	heap.Push(t, entry)
}

func (t *TTLQueue[K]) PushStd(k K) {
	if e, ok := t.keys[k]; ok {
		e.expiresAt = t.now().Add(t.ttl)
		heap.Fix(t, e.index)
		return
	}
	entry := &Entry[K]{
		k:         k,
		expiresAt: t.now().Add(t.ttl),
	}
	heap.Push(t, entry)
}

func (t *TTLQueue[K]) Push(x any) {
	t.seq++
	entry := x.(*Entry[K])
	entry.index = len(t.queue)
	entry.seq = t.seq
	t.keys[entry.k] = entry
	t.queue = append(t.queue, entry)
}

func (t *TTLQueue[K]) Pop() any {
	l := len(t.queue)
	entry := t.queue[l-1]
	t.queue[l-1] = nil
	entry.index = -1
	t.queue = t.queue[:l-1]
	delete(t.keys, entry.k)
	return entry
}

func (t *TTLQueue[K]) Peek() (*Entry[K], bool) {
	if len(t.queue) == 0 {
		return nil, false
	}
	return t.queue[0], true
}

func (t *TTLQueue[K]) Update(k K, ttl time.Duration) bool {
	entry, ok := t.keys[k]
	if !ok {
		return false
	}
	entry.expiresAt = t.now().Add(ttl)
	heap.Fix(t, entry.index)
	return true
}

func (t *TTLQueue[K]) Remove(k K) bool {
	entry, ok := t.keys[k]
	if !ok {
		return false
	}
	heap.Remove(t, entry.index)
	return true
}
