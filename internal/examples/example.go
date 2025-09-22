package main

import (
	"time"

	"github.com/jeltjongsma/go-cache/internal/cache"
	"github.com/jeltjongsma/go-cache/internal/policies"
)

func main() {
	opts := cache.NewOptions[int]().
		SetPolicy(policies.TypeLRU).
		SetCapacity(1_000_000)
	c, _ := cache.NewCache[int, string](opts)

	c.Set(1, "hello")

	val, ok := c.Get(1)
	println(val, ok) // "hello" true

	c.SetWithTTL(2, "world", 2*time.Second)
	time.Sleep(3 * time.Second)

	_, ok = c.Get(2)
	println(ok) // false (expired)
}
