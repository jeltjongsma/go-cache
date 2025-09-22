package examples

import "go-cache/internal/cache"

func main() {
	c, err := cache.NewCache[int, int]()
}
