package cache

type CacheInterface[K comparable, V any] interface {
	Set(key K, val V) bool
	Get(key K) (V, bool)
	Del(key K) bool
}

type ShardInterface[K comparable, V any] interface {
	Set(key K, val V) bool
	Get(key K) (V, bool)
	Del(key K) bool
}
