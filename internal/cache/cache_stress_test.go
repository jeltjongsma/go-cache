package cache

import (
	"go-cache/internal/context"
	"go-cache/internal/policies"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	opts := context.NewOptions[int]().
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
