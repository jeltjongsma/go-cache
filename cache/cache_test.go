package cache

import (
	"go-cache/context"
	"go-cache/policies"
	"strconv"
	"testing"
)

// might fail, but that'd mean hash function is not distributing properly
func TestCache_shardFor(t *testing.T) {
	opts := &context.Options[int]{
		Capacity:  100000,
		Policy:    nil,
		NumShards: 16,
		Hasher:    context.NewHasher[int](nil),
	}
	c, err := NewCache[int, string](
		opts,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hit := make(map[int]struct{})
	for i := range 1000000 {
		_, idx := c.shardFor(i)
		hit[int(idx)] = struct{}{}
	}

	if _, ok := hit[-1]; ok {
		t.Fatalf("expected no misses, got -1")
	}

	for i := range 16 {
		if _, ok := hit[i]; !ok {
			t.Errorf("expected idx=%d to be hit, got miss", i)
		}
	}
}

// FIXME: Can't predict all keys are actually being set, due to hash function
// Can't check if something got evicted so can't count properly still
func TestCache_Len(t *testing.T) {
	opts := &context.Options[int]{
		Capacity:  50,
		Policy:    policies.NewFIFO[int](),
		NumShards: 16,
		Hasher:    context.NewHasher[int](nil),
	}
	c, err := NewCache[int, string](opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count := 0
	for i := range 80 {
		succ, evicted := c.Set(i, strconv.Itoa(i))
		count -= evicted
		if succ {
			count++
		}
		if i%5 == 0 {
			if i < 50 {
				if c.Len() != count {
					t.Errorf("expected len=%d, got %d", count, c.Len())
				}
			} else {
				if c.Len() != count {
					t.Errorf("expected len=%d, got %d", count, c.Len())
				}
			}
		}
	}
}

func TestCache_Stats(t *testing.T) {
	opts := &context.Options[int]{
		Capacity:  5,
		Policy:    policies.NewFIFO[int](),
		NumShards: 1,
		Hasher:    context.NewHasher[int](nil),
	}
	c, err := NewCache[int, int](opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// no effect
	c.Set(1, 1)
	c.Set(2, 2)
	c.Set(3, 3)

	stats := c.Stats()
	if stats.Deletes != 0 ||
		stats.Evictions != 0 ||
		stats.Hits != 0 ||
		stats.Misses != 0 ||
		stats.Flushes != 0 {
		t.Fatalf("expected no stats, got %v", stats)
	}

	if c.Len() != 3 {
		t.Errorf("expected stats.len=3, got %d", c.Len())
	}

	// 2 hits, 1 miss
	c.Get(1)
	c.Get(3)
	c.Get(6)

	// 1 del
	c.Del(1)

	stats = c.Stats()
	if stats.Deletes != 1 ||
		stats.Evictions != 0 ||
		stats.Hits != 2 ||
		stats.Misses != 1 ||
		stats.Flushes != 0 {
		t.Fatalf("expected stats, got %v", stats)
	}

	if c.Len() != 2 {
		t.Errorf("expected stats.len=2, got %d", c.Len())
	}

	// 1 eviction
	c.Set(4, 4)
	c.Set(5, 5)
	c.Set(6, 6)
	c.Set(7, 7)

	// no effect
	c.Peek(3)

	stats = c.Stats()
	if stats.Deletes != 1 ||
		stats.Evictions != 1 ||
		stats.Hits != 2 ||
		stats.Misses != 1 ||
		stats.Flushes != 0 {
		t.Fatalf("expected stats, got %v", stats)
	}

	if c.Len() != 5 {
		t.Errorf("expected stats.len=5, got %d", c.Len())
	}

	// 1 flush, no other effects
	c.Flush()

	stats = c.Stats()
	if stats.Deletes != 1 ||
		stats.Evictions != 1 ||
		stats.Hits != 2 ||
		stats.Misses != 1 ||
		stats.Flushes != 1 {
		t.Fatalf("expected stats, got %v", stats)
	}

	if c.Len() != 0 {
		t.Errorf("expected stats.len=0, got %d", c.Len())
	}
}
