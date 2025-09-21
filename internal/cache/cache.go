package cache

// TODO: Implement tests once more features are added (e.g. hooks for logging etc.)
import (
	"errors"
	"fmt"
	"go-cache/internal/context"
	"go-cache/internal/policies"
	"runtime"
	"sync"
	"time"
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
	if opts.NumShards <= 0 {
		return nil, fmt.Errorf("num shards (%d) must be >0", opts.NumShards)
	}
	if uint64(opts.NumShards)&uint64((opts.NumShards-1)) != 0 {
		return nil, fmt.Errorf("num shards (%d) must be exponential of 2", opts.NumShards)
	}
	if opts.Hasher == nil {
		return nil, errors.New("hasher must not be nil")
	}

	// init shards
	shards := make([]*Shard[K, V], opts.NumShards)
	shardCap := opts.Capacity / int(opts.NumShards)
	for i := range opts.NumShards {
		var pol policies.Policy[K]
		switch opts.Policy {
		case policies.TypeFIFO:
			pol = policies.NewFIFO[K]()
		case policies.TypeLRU:
			pol = policies.NewLRU[K]()
		default:
			return nil, fmt.Errorf("invalid policy type: %s", opts.Policy)
		}
		shards[i] = InitShard[K, V](pol, shardCap, opts.DefaultTTL)
		StartJanitor(shards[i], 10*time.Second)
	}

	// init cache
	return &Cache[K, V]{
		shards: shards,
		hasher: opts.Hasher,
		opts:   opts,
		stats:  &Stats{},
	}, nil
}

func (c *Cache[K, V]) SetPolicy(p policies.Policy[K]) error {
	if c.Len() != 0 {
		return errors.New("cannot set policy on non-empty cache")
	}
	for _, s := range c.shards {
		s.policy = p
	}
	return nil
}

func (c *Cache[K, V]) Set(key K, val V) (success bool, evicted int) {
	shard, _ := c.shardFor(key)
	success, evicted = shard.Set(key, val)
	c.stats.Evictions.Add(uint64(evicted))
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

func (c *Cache[K, V]) shardFor(key K) (*Shard[K, V], uint64) {
	idx := c.hasher.Hash(key) % (uint64(len(c.shards)))
	return c.shards[idx], idx
}

func (c *Cache[K, V]) validate() error {
	for _, s := range c.shards {
		if err := s.validate(); err != nil {
			return err
		}
	}
	return nil
}
