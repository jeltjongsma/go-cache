package cache

// TODO: Implement tests once more features are added (e.g. hooks for logging etc.)
import (
	"fmt"
	"go-cache/context"
)

type Cache[K comparable, V any] struct {
	shards []*Shard[K, V]
	hasher *context.Hasher[K]
	opts   *context.Options[K]
}

func NewCache[K comparable, V any](
	opts *context.Options[K],
) (*Cache[K, V], error) {
	// check input
	if opts.Capacity < 0 {
		opts.Capacity = 0
	}
	if opts.NumShards&(opts.NumShards-1) != 0 {
		return nil, fmt.Errorf("num shards (%d) must be exponential of 2", opts.NumShards)
	}

	// init shards
	shards := make([]*Shard[K, V], opts.NumShards)
	shardCap := opts.Capacity / int(opts.NumShards)
	for i := range opts.NumShards {
		shards[i] = InitShard[K, V](opts.Policy, shardCap)
	}

	// init cache
	return &Cache[K, V]{
		shards: shards,
		hasher: opts.Hasher,
		opts:   opts,
	}, nil
}

func (c *Cache[K, V]) Set(key K, val V) (success bool) {
	shard, _ := c.shardFor(key)
	return shard.Set(key, val)
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	shard, _ := c.shardFor(key)
	return shard.Get(key)
}

func (c *Cache[K, V]) Del(key K) (success bool) {
	shard, _ := c.shardFor(key)
	return shard.Del(key)
}

// FIXME: Shouldn't return idx, but need it for tests for now
func (c *Cache[K, V]) shardFor(key K) (*Shard[K, V], uint64) {
	idx := c.hasher.Hash(key) % (uint64(len(c.shards)))
	return c.shards[idx], idx
}
