package cache

// TODO: Implement tests once more features are added (e.g. hooks for logging etc.)
import (
	"go-cache/context"
	"go-cache/policies"
)

type ShardedCache[K comparable, V any] struct {
	shards []*Shard[K, V]
	hasher *context.Hasher[K]
}

func NewShardedCache[K comparable, V any](
	numShards uint64,
	cap int,
	policy policies.Policy[K],
	hasher *context.Hasher[K],
) *ShardedCache[K, V] {
	shards := make([]*Shard[K, V], numShards)
	shardCap := cap / int(numShards)
	for i := range numShards {
		shards[i] = InitShard[K, V](policy, shardCap)
	}
	return &ShardedCache[K, V]{
		shards: shards,
		hasher: hasher,
	}
}

func (c *ShardedCache[K, V]) Set(key K, val V) (success bool) {
	shard, _ := c.shardFor(key)
	return shard.Set(key, val)
}

func (c *ShardedCache[K, V]) Get(key K) (V, bool) {
	shard, _ := c.shardFor(key)
	return shard.Get(key)
}

func (c *ShardedCache[K, V]) Del(key K) (success bool) {
	shard, _ := c.shardFor(key)
	return shard.Del(key)
}

// FIXME: Shouldn't return idx, but need it for tests for now
func (c *ShardedCache[K, V]) shardFor(key K) (*Shard[K, V], uint64) {
	idx := c.hasher.Hash(key) % (uint64(len(c.shards)))
	return c.shards[idx], idx
}
