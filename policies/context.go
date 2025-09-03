package policies

type Policy[K comparable] interface {
	OnHit(key K)
	OnSet(key K)
	OnDel(key K)
	Evict() (K, bool)
	Type() PolicyType
}

type PolicyType int

const (
	TypeFIFO PolicyType = iota
)
