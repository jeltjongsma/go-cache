# Go cache
A simple concurrent, sharded, in-memory cache written in Go. 
Built to explore caching internals, eviction policies, TTL management, and performance trade-offs.

## Features

- Sharded cache to reduce lock contention
- Configurable cache (eviction policy, number of shards, TTL, etc.)
- Support for per-entry TTLs
- Concurrent safe via sharded locks
- Thread-safe stats tracking (hits, misses, evictions, etc.)
- Background janitor to reduce stale entries

## API overview
- `NewOptions[K comparable]() *Options[K]` - returns default options. 
Setters return `*Options[K]`, so they can be chained:
    - `.SetCapacity(c int)` - set to 0 for no limit
    - `.SetPolicy(p PolicyType)` 
    - `.SetNumShards(n int)` 
    - `.SetHasher(h *Hasher[K])` 
    - `.SetDefaultTTL(ttl time.Duration)` - set to 0 for no expiration
- `NewCache[K comparable, V any](*Options[K]) (*Cache[K, V], error)`
- `.Set(key K, val V) (success bool, evicted int)` 
- `.SetWithTTL(key K, val V, ttl time.Duration) (success bool, evicted int)`
- `.Get(key K) (val V, hit bool)`
- `.Peek(key K) (V, bool)` - read without policy effects
- `.Del(key K) (success bool)` - remove key
- `.Len() int` - number of keys stored
- `.Flush()` - clear cache
- `.Stats() *StatsSnapshot` - counters for hits, misses, evictions, deletes, and flushes
- `.SetPolicy(Policy[K]) error` - set custom policy

## Usage

### Installation
```bash
go get github.com/jeltjongsma/go-cache
```

### Example
*Note: The import path is github.com/jeltjongsma/go-cache, but the package name is cache*
```golang
package main

import (
	"fmt"
	"time"

	"github.com/jeltjongsma/go-cache"
	"github.com/jeltjongsma/go-cache/pkg/policies"
)

func main() {
	opts := cache.NewOptions[int]().
		SetPolicy(policies.TypeLRU).
		SetCapacity(1_000_000)
	c, _ := cache.NewCache[int, string](opts)

	c.Set(1, "hello")

	val, ok := c.Get(1)
	fmt.Println(val, ok)

	c.SetWithTTL(2, "world", 2*time.Second)
	time.Sleep(3 * time.Second)

	_, ok = c.Get(2) // expired
	fmt.Println(ok)
}
```
**Example output:**
```text
hello true
false
```

### Testing
```bash
go test ./...
```

### Benchmarks
```bash
go test ./... -bench=. -benchmem
```

## Ideas for future work
- Implement support for callbacks (e.g., `(*Options).OnEvict(k K, victim V)`) to allow for logging, metrics, etc.
- Implement more eviction policies (e.g., LFU, ARC)
- Improve stats tracking (e.g., expirations, hot keys)
- Improve performance:
	- Increase benchmark coverage
	- Explore optimal number of shards
	- Optimize common operations (`Get`, `Set`) 

## License
This project is licensed under the [MIT license](LICENSE)

## Contributing
This project started as a way to explore caching internals in Go, so discussions, bug reports, and feature suggestions are especially valuable.  

If you’d like to contribute:
- Open an issue to discuss new ideas, improvements, or bugs
- Fork the repo and create a feature branch (`git checkout -b feat/my-feature`)
- Run `go test ./...` to ensure all tests pass
- Submit a pull request with a clear description of the changes
