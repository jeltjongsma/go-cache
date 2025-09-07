package cache

import (
	"go-cache/context"
	"go-cache/policies"
	"sync"
)

type Cache[K comparable, V any] struct {
	mu     sync.RWMutex
	opts   *context.Options[K]
	store  map[K]V
	policy policies.Policy[K]
}

func New[K comparable, V any](opts *context.Options[K]) *Cache[K, V] {
	return &Cache[K, V]{
		opts:   opts,
		store:  make(map[K]V, opts.Capacity),
		policy: opts.Policy,
	}
}

func (c *Cache[K, V]) Set(key K, val V) (success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, exists := c.store[key]

	// if capacity = 0 the cache is considered unbounded
	if !exists && c.opts.Capacity > 0 {
		attempts := 0
		for len(c.store) >= c.opts.Capacity {
			victim, ok := c.policy.Evict()
			if !ok {
				return false // refuse insert
			}
			if _, present := c.store[victim]; present {
				delete(c.store, victim)
			} else {
				// policy and store out of sync
				attempts++
				if attempts > c.opts.Capacity+1 {
					return false // refuse insert on too many eviction attempts
				}
			}
		}
	}

	c.store[key] = val
	c.policy.OnSet(key)
	return true
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	val, ok := c.store[key]
	if !ok { // cache miss
		var zero V
		return zero, false
	}
	c.policy.OnHit(key)
	return val, true
}

func (c *Cache[K, V]) Del(key K) (success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.store[key]; !ok {
		return false
	}
	delete(c.store, key)
	c.policy.OnDel(key)
	return true
}
