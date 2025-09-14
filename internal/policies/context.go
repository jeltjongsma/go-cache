package policies

import "reflect"

type Policy[K comparable] interface {
	Type() (PolicyType, reflect.Type)
	OnHit(K)
	OnSet(K)
	OnDel(K)
	Evict() (K, bool)
	Reset()
	Equals(Policy[any]) bool
}

type PolicyType string

const (
	TypeFIFO PolicyType = "FIFO"
	TypeLRU  PolicyType = "LRU"
)
