package ttl_queue

import (
	"testing"
	"time"
)

func TestNewTTLQueue(t *testing.T) {
	q := NewTTLQueue[int](500)
	if q.ttl != 500 {
		t.Errorf("expected ttl=500, got %d", q.ttl)
	}
	if q.keys == nil {
		t.Errorf("expected keys!=nil, got nil")
	}
	if q.seq != 0 {
		t.Errorf("expected seq=0, got %d", q.seq)
	}
}

func TestLen(t *testing.T) {
	q := NewTTLQueue[int](100)
	q.queue = []*Entry[int]{{}, {}, {}, {}, {}}
	if l := q.Len(); l != 5 {
		t.Errorf("expected len=5, got %d", l)
	}
}

func TestLess(t *testing.T) {
	q := NewTTLQueue[int](100)
	n := time.Now()
	q.now = func() time.Time {
		return n
	}
	q.queue = []*Entry[int]{
		{expiresAt: q.now(), seq: 0},
		{expiresAt: q.now(), seq: 1},
		{expiresAt: q.now().Add(50), seq: 0},
	}

	if q.Less(0, 1) != true {
		t.Errorf("expected true, got false")
	}
	if q.Less(0, 2) != true {
		t.Errorf("expected true, got false")
	}
	if q.Less(1, 2) != true {
		t.Errorf("expected true, got false")
	}
}

func TestSwap(t *testing.T) {
	q := NewTTLQueue[int](100)
	q.queue = []*Entry[int]{{K: 1, index: 0}, {K: 2, index: 1}}

	q.Swap(0, 1)

	if x := q.queue[0]; x.K != 2 || x.index != 0 {
		t.Errorf("expected k=2 & idx=0, got %d, %d", x.K, x.index)
	}
	if x := q.queue[1]; x.K != 1 || x.index != 1 {
		t.Errorf("expected k=1 & idx=1, got %d, %d", x.K, x.index)
	}
}

func TestPush(t *testing.T) {
	q := NewTTLQueue[int](100)
	n := time.Now()
	q.now = func() time.Time {
		return n
	}
	e := &Entry[int]{
		K:         1,
		expiresAt: q.now(),
	}

	q.Push(e)

	if len(q.queue) != 1 {
		t.Fatalf("expected len=1, got %d", len(q.queue))
	}
	got := q.queue[0]
	if got.seq != 1 {
		t.Errorf("expected seq=1, got %d", got.seq)
	}
	if got.index != 0 {
		t.Errorf("expected idx=0, got %d", got.index)
	}
	if got.K != 1 {
		t.Errorf("expected k=1, got %d", got.K)
	}
	if !got.expiresAt.Equal(e.expiresAt) {
		t.Errorf("got different expiresAt's")
	}
	if _, ok := q.keys[1]; !ok {
		t.Errorf("k=1 not in map")
	}
}

func TestPushWithTTL_NewEntry(t *testing.T) {
	q := NewTTLQueue[int](100)
	n := time.Now()
	q.now = func() time.Time {
		return n
	}

	q.PushWithTTL(1, 50)

	if len(q.queue) != 1 {
		t.Fatalf("expected len=1, got %d", len(q.queue))
	}
	got := q.queue[0]
	if got.K != 1 {
		t.Errorf("expected k=1, got %d", got.K)
	}
	if !got.expiresAt.Equal(n.Add(50)) {
		t.Errorf("expected ttl + 50, got %v", got.expiresAt)
	}
}

func TestPushWithTTL_Duplicate(t *testing.T) {
	q := NewTTLQueue[int](100)
	n := time.Now()
	q.now = func() time.Time {
		return n
	}

	q.PushStd(1)
	q.PushWithTTL(1, 50)

	if len(q.queue) != 1 {
		t.Fatalf("expected len=1, got %d", len(q.queue))
	}
	got := q.queue[0]
	if got.K != 1 {
		t.Errorf("expected k=1, got %d", got.K)
	}
	if !got.expiresAt.Equal(n.Add(50)) {
		t.Errorf("expected ttl + 50, got %v", got.expiresAt)
	}
}

func TestPushStd_NewEntry(t *testing.T) {
	q := NewTTLQueue[int](100)
	n := time.Now()
	q.now = func() time.Time {
		return n
	}

	q.PushStd(1)

	if len(q.queue) != 1 {
		t.Fatalf("expected len=1, got %d", len(q.queue))
	}
	got := q.queue[0]
	if got.K != 1 {
		t.Errorf("expected k=1, got %d", got.K)
	}
	if !got.expiresAt.Equal(n.Add(100)) {
		t.Errorf("expected ttl + 100, got %v", got.expiresAt)
	}
}

func TestPushStd_Duplicate(t *testing.T) {
	q := NewTTLQueue[int](100)
	n := time.Now()
	q.now = func() time.Time {
		return n
	}

	q.PushStd(1)
	q.PushStd(1)

	if len(q.queue) != 1 {
		t.Fatalf("expected len=1, got %d", len(q.queue))
	}
	got := q.queue[0]
	if got.K != 1 {
		t.Errorf("expected k=1, got %d", got.K)
	}
	if !got.expiresAt.Equal(n.Add(100)) {
		t.Errorf("expected ttl + 100, got %v", got.expiresAt)
	}
}

func TestPopMin(t *testing.T) {
	q := NewTTLQueue[int](100)
	n := time.Now()
	q.now = func() time.Time {
		return n
	}

	q.PushStd(1)
	e := q.PopMin()

	if e.K != 1 {
		t.Errorf("expected k=1, got %d", e.K)
	}
	if len(q.queue) != 0 {
		t.Errorf("expected len=0, got %d", len(q.queue))
	}
	if _, ok := q.keys[1]; ok {
		t.Errorf("k=1 shouldn't be in map")
	}
	if e.index != -1 {
		t.Errorf("expected index=-1, got %d", e.index)
	}
}

func TestPeek(t *testing.T) {
	q := NewTTLQueue[int](100)
	n := time.Now()
	q.now = func() time.Time {
		return n
	}

	if _, ok := q.Peek(); ok {
		t.Fatalf("expected false, got true")
	}

	q.PushStd(1)
	e, ok := q.Peek()
	if !ok {
		t.Fatalf("expected true, got false")
	}
	if e.K != 1 {
		t.Errorf("expected k=1, got %d", e.K)
	}
	if len(q.queue) != 1 {
		t.Errorf("expected len=1, got %d", len(q.queue))
	}
	if _, ok := q.keys[1]; !ok {
		t.Errorf("k=1 not in map")
	}
	if e.index != 0 {
		t.Errorf("expected index=0, got %d", e.index)
	}
}

func TestHasExpired(t *testing.T) {
	q := NewTTLQueue[int](100)
	n := time.Now()
	q.now = func() time.Time {
		return n
	}

	q.queue = []*Entry[int]{
		{expiresAt: n.Add(-50)},
		{expiresAt: n},
		{expiresAt: n.Add(50)},
	}

	if !q.HasExpired() {
		t.Fatalf("expected true, got false")
	}
	q.PopMin()
	if !q.HasExpired() {
		t.Fatalf("expected true, got false")
	}
	q.PopMin()
	if q.HasExpired() {
		t.Fatalf("expected false, got true")
	}
	q.PopMin()
	if q.HasExpired() {
		t.Fatalf("expected false, got true")
	}
}

func TestUpdate(t *testing.T) {
	q := NewTTLQueue[int](100)
	n := time.Now()
	q.now = func() time.Time {
		return n
	}

	if ok := q.Update(1, 50); ok {
		t.Fatalf("expected false, got true")
	}

	q.PushStd(1)

	if ok := q.Update(1, 50); !ok {
		t.Fatalf("expected true, got false")
	}
	got, _ := q.Peek()
	if !got.expiresAt.Equal(n.Add(50)) {
		t.Errorf("expected ttl + 50, got %v", got.expiresAt)
	}
	if got.seq != q.seq {
		t.Errorf("expected seq=%d, got %d", q.seq, got.seq)
	}
}

func TestRemove(t *testing.T) {
	q := NewTTLQueue[int](100)
	n := time.Now()
	q.now = func() time.Time {
		return n
	}

	if ok := q.Remove(1); ok {
		t.Fatalf("expected false, got true")
	}

	q.PushStd(1)

	if ok := q.Remove(1); !ok {
		t.Fatalf("expected true, got false")
	}
	if l := len(q.queue); l != 0 {
		t.Errorf("expected len=0, got %d", l)
	}
	if _, ok := q.keys[1]; ok {
		t.Errorf("k=1 shouldn't be in map")
	}
}

func TestReset(t *testing.T) {
	q := NewTTLQueue[int](100)

	q.PushStd(1)
	q.PushStd(2)
	q.PushStd(3)

	q.Reset()

	if l := len(q.queue); l != 0 {
		t.Errorf("expected len=0, got %d", l)
	}
	if l := len(q.keys); l != 0 {
		t.Errorf("expected len=0, got %d", l)
	}
	if q.seq != 0 {
		t.Errorf("expected seq=0, got %d", q.seq)
	}
}

func TestTTLQueue(t *testing.T) {
	q := NewTTLQueue[int](100)
	n := time.Now()
	q.now = func() time.Time {
		return n
	}

	// TTL based
	for i := range 50 {
		q.PushWithTTL(i, time.Duration(i))
	}

	for i := range 50 {
		e := q.PopMin()
		if e.K != i {
			t.Errorf("expected k=%d, got %d", i, e.K)
		}
	}

	q.Reset()

	// seq based
	for i := range 50 {
		q.PushStd(i)
	}

	for i := range 50 {
		if e := q.PopMin(); e.K != i {
			t.Errorf("expected k=%d, got %d", i, e.K)
		}
	}

	q.Reset()

	// after updates
	for i := range 20 {
		q.PushStd(i)
	}

	for i := range 10 {
		q.Update(i, time.Duration(i))
	}

	for i := range 10 {
		if e := q.PopMin(); e.K != i {
			t.Errorf("expected k=%d, got %d", i, e.K)
		}
	}

	for i := range 10 {
		if e := q.PopMin(); e.K != i+10 {
			t.Errorf("expected k=%d, got %d", i+10, e.K)
		}
	}
}

func TestSetNow(t *testing.T) {
	q := NewTTLQueue[int](100)
	n := time.Now()
	q.SetNow(func() time.Time { return n })

	if !q.now().Equal(n) {
		t.Errorf("now not set")
	}
}
