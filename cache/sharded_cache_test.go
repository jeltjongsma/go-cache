package cache

import (
	"go-cache/context"
	"testing"
)

// might fail, but that'd mean hash function is distributing properly
func TestCache_shardFor(t *testing.T) {
	c := NewShardedCache[int, string](
		16,
		100000,
		nil,
		context.NewHasher[int](nil),
	)

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
