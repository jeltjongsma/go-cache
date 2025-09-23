package ttl_queue

import (
	"container/heap"
	"time"
)

type Entry[K comparable] struct {
	ExpiresAt time.Time
	index     int
	seq       uint64
	K         K
}

type TTLQueue[K comparable] struct {
	queue []*Entry[K]
	ttl   time.Duration
	keys  map[K]*Entry[K]
	now   func() time.Time
	seq   uint64
}

// TTLQueue is an expiration based priority queue.
// The queue is ordered by (ExpiresAt, seq) ascending.
func NewTTLQueue[K comparable](defaultTTL time.Duration) *TTLQueue[K] {
	q := &TTLQueue[K]{
		ttl:  defaultTTL,
		keys: make(map[K]*Entry[K]),
		now:  time.Now,
		seq:  0,
	}
	heap.Init(q)
	return q
}

func (t *TTLQueue[K]) Len() int {
	return len(t.queue)
}

func (t *TTLQueue[K]) Less(i, j int) bool {
	x, y := t.queue[i], t.queue[j]
	if x.ExpiresAt.Equal(y.ExpiresAt) {
		return x.seq < y.seq
	}
	return x.ExpiresAt.Before(y.ExpiresAt)
}

func (t *TTLQueue[K]) Swap(i, j int) {
	t.queue[i], t.queue[j] = t.queue[j], t.queue[i]
	t.queue[i].index = i
	t.queue[j].index = j
}

// PushWithTTL pushes a new entry onto the heap with a given TTL.
// If the entry already exists the TTL will be updated and get bubbled up.
func (t *TTLQueue[K]) PushWithTTL(k K, ttl time.Duration) {
	if e, ok := t.keys[k]; ok {
		e.ExpiresAt = t.now().Add(ttl)
		t.seq++
		e.seq = t.seq
		heap.Fix(t, e.index)
		return
	}
	entry := &Entry[K]{
		K:         k,
		ExpiresAt: t.now().Add(ttl),
	}
	heap.Push(t, entry)
}

// PushStd pushes a new entry onto the heap with the default TTL.
// If the entry already exists the TTL will be updated and get bubbled up.
func (t *TTLQueue[K]) PushStd(k K) {
	if e, ok := t.keys[k]; ok {
		e.ExpiresAt = t.now().Add(t.ttl)
		t.seq++
		e.seq = t.seq
		heap.Fix(t, e.index)
		return
	}
	entry := &Entry[K]{
		K:         k,
		ExpiresAt: t.now().Add(t.ttl),
	}
	heap.Push(t, entry)
}

// Push is called by container/heap (never use yourself).
func (t *TTLQueue[K]) Push(x any) {
	entry := x.(*Entry[K])
	t.seq++
	entry.seq = t.seq
	entry.index = len(t.queue)
	t.keys[entry.K] = entry
	t.queue = append(t.queue, entry)
}

// PopMin pops the entry that expires earliest.
func (t *TTLQueue[K]) PopMin() *Entry[K] {
	return heap.Pop(t).(*Entry[K])
}

// Pop is called by container/heap (never use yourself).
func (t *TTLQueue[K]) Pop() any {
	l := len(t.queue)
	entry := t.queue[l-1]
	t.queue[l-1] = nil
	entry.index = -1
	t.queue = t.queue[:l-1]
	delete(t.keys, entry.K)
	return entry
}

// Peek returns the entry at the front of the queue (earliest expiration).
func (t *TTLQueue[K]) Peek() (*Entry[K], bool) {
	if len(t.queue) == 0 {
		return nil, false
	}
	return t.queue[0], true
}

// HasExpired returns whether the first entry of the queue is expired.
func (t *TTLQueue[K]) HasExpired() bool {
	if len(t.queue) == 0 {
		return false
	}
	e := t.queue[0]
	if e.ExpiresAt.Before(t.now()) || e.ExpiresAt.Equal(t.now()) {
		return true
	}
	return false
}

func (t *TTLQueue[K]) Update(k K, ttl time.Duration) bool {
	entry, ok := t.keys[k]
	if !ok {
		return false
	}
	entry.ExpiresAt = t.now().Add(ttl)
	t.seq++
	entry.seq = t.seq
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

func (t *TTLQueue[K]) Reset() {
	for i := range t.queue {
		t.queue[i] = nil
	}
	t.queue = t.queue[:0]
	clear(t.keys)
	t.seq = 0
}

// SetNow sets the internal `now` function.
// Useful for deterministic tests.
func (t *TTLQueue[K]) SetNow(now func() time.Time) {
	t.now = now
}
