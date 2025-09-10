package cache

// TODO: Implement tests once more features are added (e.g. hooks for logging etc.)
import (
	"fmt"
	"go-cache/context"
	"runtime"
	"sync"
)

type Cache[K comparable, V any] struct {
	shards []*Shard[K, V]
	hasher *context.Hasher[K]
	opts   *context.Options[K]
	stats  *Stats
}

func NewCache[K comparable, V any](
	opts *context.Options[K],
) (*Cache[K, V], error) {
	// check input
	if opts.Capacity < 0 {
		opts.Capacity = 0
	}
	if uint64(opts.NumShards)&uint64((opts.NumShards-1)) != 0 {
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
		stats:  &Stats{},
	}, nil
}

func (c *Cache[K, V]) Set(key K, val V) (success bool, evicted int) {
	shard, _ := c.shardFor(key)
	success, evicted = shard.Set(key, val)
	if evicted > 0 {
		c.stats.Evictions.Add(1)
	}
	return
}

func (c *Cache[K, V]) Get(key K) (val V, hit bool) {
	shard, _ := c.shardFor(key)
	val, hit = shard.Get(key)
	if hit {
		c.stats.Hits.Add(1)
	} else {
		c.stats.Misses.Add(1)
	}
	return
}

func (c *Cache[K, V]) Peek(key K) (V, bool) {
	shard, _ := c.shardFor(key)
	return shard.Peek(key)
}

func (c *Cache[K, V]) Del(key K) (success bool) {
	shard, _ := c.shardFor(key)
	if success = shard.Del(key); success {
		c.stats.Deletes.Add(1)
	}
	return
}

func (c *Cache[K, V]) Len() int {
	sum := 0
	for _, s := range c.shards {
		sum += len(s.store)
	}
	return sum
}

func (c *Cache[K, V]) Flush() {
	numWorkers := min(runtime.GOMAXPROCS(0), c.opts.NumShards)
	jobs := make(chan int, numWorkers)
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for w := 0; w < numWorkers; w++ {
		go func() {
			defer wg.Done()
			for i := range jobs {
				s := c.shards[i]
				s.Flush()
			}
		}()
	}

	for i := range c.opts.NumShards {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	c.stats.Flushes.Add(1)
}

func (c *Cache[K, V]) Stats() *StatsSnapshot {
	return &StatsSnapshot{
		Hits:      c.stats.Hits.Load(),
		Misses:    c.stats.Misses.Load(),
		Evictions: c.stats.Evictions.Load(),
		Deletes:   c.stats.Deletes.Load(),
		Flushes:   c.stats.Flushes.Load(),
	}
}

// FIXME: Shouldn't return idx, but need it for tests for now
func (c *Cache[K, V]) shardFor(key K) (*Shard[K, V], uint64) {
	idx := c.hasher.Hash(key) % (uint64(len(c.shards)))
	return c.shards[idx], idx
}
