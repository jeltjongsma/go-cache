package cache

import (
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/jeltjongsma/go-cache/pkg/policies"
)

func TestCache(t *testing.T) {
	opts := NewOptions[int]().
		SetCapacity(100_000).
		SetPolicy(policies.TypeLRU).
		SetNumShards(128)
	c, err := NewCache[int, int](opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	numWorkers := runtime.GOMAXPROCS(0)
	jobs := make(chan int, numWorkers)
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for w := 0; w < numWorkers; w++ {
		go func(seed int64) {
			defer wg.Done()
			r := rand.New(rand.NewSource(seed))
			for i := range jobs {
				k := r.Intn(10_000)
				switch r.Intn(6) {
				case 0, 1, 2:
					c.Set(k, r.Int())
				case 3, 4:
					c.Get(k)
				case 5:
					c.Del(k)
				}
				if i%10_000 == 0 {
					if err := c.validate(); err != nil {
						panic(err)
					}
				}
			}
		}(rng.Int63())
	}

	for i := range 1_000_000 {
		jobs <- i
	}
	close(jobs)
}

type mm[K comparable, V any] struct {
	store map[K]V
	mu    sync.RWMutex
}

func BenchmarkBaselineSet(b *testing.B) {
	bl := mm[int, int]{store: make(map[int]int, b.N)}
	var k int
	b.ResetTimer()
	for i := range b.N {
		k++
		bl.mu.Lock()
		bl.store[k] = i
		bl.mu.Unlock()
	}
}

var sink int

func BenchmarkBaselineGet(b *testing.B) {
	bl := mm[int, int]{store: make(map[int]int, b.N)}

	const N = 1 << 16
	for i := range N {
		bl.store[i] = i
	}

	mask := N - 1

	b.ResetTimer()
	for i := range b.N {
		sink = bl.store[i%mask]
	}
}

func BenchmarkSet(b *testing.B) {
	opts := NewOptions[int]().
		SetCapacity(b.N).
		SetPolicy(policies.TypeLRU).
		SetNumShards(128)
	c, err := NewCache[int, int](opts)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	b.ResetTimer()
	for i := range b.N {
		c.Set(i, i)
	}
}

func BenchmarkGet(b *testing.B) {
	opts := NewOptions[int]().
		SetCapacity(b.N).
		SetPolicy(policies.TypeLRU).
		SetNumShards(128).
		SetDefaultTTL(10)
	c, err := NewCache[int, int](opts)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	const N = 1 << 16
	for i := range N {
		if i%4 != 0 { // 1 in 4 miss
			c.Set(i, i)
		}
	}

	mask := N - 1

	b.ResetTimer()
	for i := range b.N {
		c.Get(i % mask)
	}
}
