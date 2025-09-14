package cache

import (
	"sync/atomic"
)

type CacheInterface[K comparable, V any] interface {
	Set(key K, val V) (success bool, evicted int)
	Get(key K) (V, bool)
	Peek(key K) (V, bool)
	Del(key K) bool
	Len() int
	Flush()
	Stats() *StatsSnapshot
}

type Stats struct {
	Hits      atomic.Uint64
	Misses    atomic.Uint64
	Evictions atomic.Uint64
	Deletes   atomic.Uint64
	Flushes   atomic.Uint64
}

type StatsSnapshot struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
	Deletes   uint64
	Flushes   uint64
}
