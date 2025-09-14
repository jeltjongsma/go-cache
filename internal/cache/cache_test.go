package cache

import (
	"go-cache/internal/context"
	"go-cache/internal/policies"
	"testing"
)

func TestCache_New(t *testing.T) {
	tests := []struct {
		name    string
		opts    *context.Options[int]
		wantErr bool
	}{
		{"no error", &context.Options[int]{
			Capacity:  2,
			Policy:    policies.TypeFIFO,
			NumShards: 2,
			Hasher:    context.NewHasher[int](nil),
		}, false},
		{"negative cap", &context.Options[int]{
			Capacity:  -1,
			Policy:    policies.TypeFIFO,
			NumShards: 2,
			Hasher:    context.NewHasher[int](nil),
		}, false},
		{"incorrect num shards", &context.Options[int]{
			Capacity:  2,
			Policy:    policies.TypeFIFO,
			NumShards: 3,
			Hasher:    context.NewHasher[int](nil),
		}, true},
		{"nil policy", &context.Options[int]{
			Capacity:  2,
			Policy:    "wrong policy",
			NumShards: 2,
			Hasher:    context.NewHasher[int](nil),
		}, true},
		{"nil hasher", &context.Options[int]{
			Capacity:  2,
			Policy:    policies.TypeFIFO,
			NumShards: 2,
			Hasher:    nil,
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCache[int, int](tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tt.opts.Capacity < 0 {
					if c.opts.Capacity != 0 {
						t.Errorf("expected cap=0, got %d", c.opts.Capacity)
					}
				} else {
					if c.opts.Capacity != tt.opts.Capacity {
						t.Errorf("expected c.cap==tt.cap, got %d!=%d", c.opts.Capacity, tt.opts.Capacity)
					}
				}
				if len(c.shards) != 2 {
					t.Fatalf("expected len(c.shards)=2, got %d", len(c.shards))
				}
				for i, s := range c.shards {
					if s == nil {
						t.Fatalf("shard %d is nil", i)
					}
				}
			}
		})
	}
}

func TestCache_Set(t *testing.T) {
	c, err := NewCache[int, int](context.NewOptions[int]().
		SetCapacity(100).
		SetPolicy(policies.TypeFIFO).
		SetNumShards(16).
		SetHasher(context.NewHasher[int](nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	total_evicted := 0
	for i := range 1000 {
		success, evicted := c.Set(i, i)
		if !success {
			t.Fatalf("expected true, got false: %d", i)
		}
		total_evicted += evicted
	}
	if stats := c.Stats(); stats.Evictions != uint64(total_evicted) {
		t.Fatalf("expected stats.evicted=total_evicted, got %d!=%d", stats.Evictions, total_evicted)
	}
}

func TestCache_Get(t *testing.T) {
	c, err := NewCache[int, int](context.NewOptions[int]().
		SetCapacity(100).
		SetPolicy(policies.TypeFIFO).
		SetNumShards(16).
		SetHasher(context.NewHasher[int](nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := range 50 {
		c.Set(i, i)
	}

	total_hit := 0
	total_miss := 0
	var zero int
	for i := range 100 {
		val, hit := c.Get(i)
		if hit {
			total_hit++
		} else {
			total_miss++
			if val != zero {
				t.Errorf("expected zero value, got %d", val)
			}
		}
	}

	stats := c.Stats()
	if stats.Hits != uint64(total_hit) {
		t.Errorf("expected %d hits, got %d", total_hit, stats.Hits)
	}
	if stats.Misses != uint64(total_miss) {
		t.Errorf("expected %d misses, got %d", total_miss, stats.Misses)
	}
}

func TestCache_Peek(t *testing.T) {
	c, err := NewCache[int, int](context.NewOptions[int]().
		SetCapacity(100).
		SetPolicy(policies.TypeFIFO).
		SetNumShards(16).
		SetHasher(context.NewHasher[int](nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := range 50 {
		c.Set(i, i)
	}

	for i := range 100 {
		c.Peek(i)
	}

	stats := c.Stats()
	if stats.Hits != 0 {
		t.Errorf("expected 0 hits, got %d", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("expected 0 misses, got %d", stats.Misses)
	}
}

func TestCache_Del(t *testing.T) {
	c, err := NewCache[int, int](context.NewOptions[int]().
		SetCapacity(100).
		SetPolicy(policies.TypeFIFO).
		SetNumShards(16).
		SetHasher(context.NewHasher[int](nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := range 50 {
		c.Set(i, i)
	}

	total_del := 0
	for i := range 100 {
		if ok := c.Del(i); ok {
			total_del++
		}
	}

	stats := c.Stats()
	if stats.Deletes != uint64(total_del) {
		t.Errorf("expected %d deletes, got %d", total_del, stats.Deletes)
	}
}

func TestCache_Flush(t *testing.T) {
	c, err := NewCache[int, int](context.NewOptions[int]().
		SetCapacity(100).
		SetPolicy(policies.TypeFIFO).
		SetNumShards(16).
		SetHasher(context.NewHasher[int](nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := range 100 {
		c.Set(i, i)
	}

	c.Flush()

	for i := range 100 {
		_, ok := c.Peek(i)
		if ok {
			t.Fatalf("expected false, got true")
		}
	}

	stats := c.Stats()
	if stats.Flushes != 1 {
		t.Errorf("expected 1 deletes, got %d", stats.Flushes)
	}
}

// might fail, but that'd mean hash function is not distributing properly
func TestCache_shardFor(t *testing.T) {
	c, err := NewCache[int, int](context.NewOptions[int]().
		SetCapacity(10000).
		SetPolicy(policies.TypeFIFO).
		SetNumShards(16).
		SetHasher(context.NewHasher[int](nil)))
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
	c, err := NewCache[int, int](context.NewOptions[int]().
		SetCapacity(50).
		SetPolicy(policies.TypeFIFO).
		SetNumShards(16).
		SetHasher(context.NewHasher[int](nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count := 0
	for i := range 80 {
		succ, evicted := c.Set(i, i)
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
	c, err := NewCache[int, int](context.NewOptions[int]().
		SetCapacity(5).
		SetPolicy(policies.TypeFIFO).
		SetNumShards(1).
		SetHasher(context.NewHasher[int](nil)))
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
