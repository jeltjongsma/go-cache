package policies

import "reflect"

type Policy[K comparable] interface {
	OnHit(key K)
	OnSet(key K)
	OnDel(key K)
	Evict() (K, bool)
	Type() (PolicyType, reflect.Type)
}

type PolicyType string

const (
	TypeFIFO PolicyType = "FIFO"
	TypeLRU  PolicyType = "LRU"
)
