package ttl_queue

import (
	"container/heap"
	"time"
)

type Entry[K comparable] struct {
	K         K
	expiresAt time.Time
	index     int
	seq       uint64
}

// Invariants:
//   - keys maps each K present in queue to its Entry.
//   - entry.index is the current position in queue, or -1 when not present.
//   - Less orders by (expiresAt, seq) ascending ⇒ min-heap.
type TTLQueue[K comparable] struct {
	queue []*Entry[K]
	ttl   time.Duration
	keys  map[K]*Entry[K]
	now   func() time.Time
	seq   uint64
}

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
		t.seq++
		e.seq = t.seq
		heap.Fix(t, e.index)
		return
	}
	entry := &Entry[K]{
		K:         k,
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
		K:         k,
		expiresAt: t.now().Add(t.ttl),
	}
	heap.Push(t, entry)
}

// called by container/heap (never use yourself)
func (t *TTLQueue[K]) Push(x any) {
	t.seq++
	entry := x.(*Entry[K])
	entry.index = len(t.queue)
	entry.seq = t.seq
	t.keys[entry.K] = entry
	t.queue = append(t.queue, entry)
}

func (t *TTLQueue[K]) PopMin() *Entry[K] {
	return heap.Pop(t).(*Entry[K])
}

// called by container/heap (never use yourself)
func (t *TTLQueue[K]) Pop() any {
	l := len(t.queue)
	entry := t.queue[l-1]
	t.queue[l-1] = nil
	entry.index = -1
	t.queue = t.queue[:l-1]
	delete(t.keys, entry.K)
	return entry
}

func (t *TTLQueue[K]) Peek() (*Entry[K], bool) {
	if len(t.queue) == 0 {
		return nil, false
	}
	return t.queue[0], true
}

func (t *TTLQueue[K]) HasExpired() bool {
	if len(t.queue) == 0 {
		return false
	}
	e := t.queue[0]
	if e.expiresAt.Before(t.now()) || e.expiresAt.Equal(t.now()) {
		return true
	}
	return false
}

func (t *TTLQueue[K]) Update(k K, ttl time.Duration) bool {
	entry, ok := t.keys[k]
	if !ok {
		return false
	}
	entry.expiresAt = t.now().Add(ttl)
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

func (t *TTLQueue[K]) SetNow(now func() time.Time) {
	t.now = now
}
