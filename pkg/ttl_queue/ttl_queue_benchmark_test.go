package ttl_queue

import (
	"testing"
	"time"
)

func BenchmarkTTLQueue_PushStd(b *testing.B) {
	q := NewTTLQueue[int](300_000)

	b.ResetTimer()
	for i := range b.N {
		q.PushStd(i)
	}
}

func BenchmarkTTLQueue_PushWithTTL(b *testing.B) {
	q := NewTTLQueue[int](300_000)

	b.ResetTimer()
	for i := range b.N {
		q.PushStd(i)
	}
}

func BenchmarkTTLQueue_Swap(b *testing.B) {
	q := NewTTLQueue[int](300_000)

	q.PushStd(1)
	q.PushStd(2)

	b.ResetTimer()
	for range b.N {
		q.Swap(0, 1)
	}
}

func BenchmarkTTLQueue_Less(b *testing.B) {
	q := NewTTLQueue[int](300_000)

	q.PushStd(1)
	q.PushStd(2)

	b.ResetTimer()
	for range b.N {
		q.Less(0, 1)
	}
}

func BenchmarkTTLQueue_Push(b *testing.B) {
	q := NewTTLQueue[int](300_000)
	n := time.Now()
	q.SetNow(func() time.Time { return n })

	const N = 1 << 16
	entries := make([]Entry[int], N)
	for i := range N {
		entries[i] = Entry[int]{K: i, ExpiresAt: n.Add(300_000), index: i, seq: uint64(i)}
	}

	mask := N - 1

	b.ResetTimer()
	for i := range b.N {
		q.Push(&entries[i&mask])
	}
}
