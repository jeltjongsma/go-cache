package cache

import (
	"go-cache/context"
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
		t.Fatalf("unexpected error")
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
