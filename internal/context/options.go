package context

import (
	"go-cache/internal/policies"
	"time"
)

type Options[K comparable] struct {
	Capacity   int
	Policy     policies.PolicyType
	NumShards  int
	Hasher     *Hasher[K]
	DefaultTTL time.Duration
}

func NewOptions[K comparable]() *Options[K] {
	return &Options[K]{
		Capacity:   100,
		Policy:     policies.TypeFIFO,
		NumShards:  16,
		Hasher:     NewHasher[K](nil),
		DefaultTTL: 300_000,
	}
}

func (o *Options[K]) SetCapacity(c int) *Options[K] {
	o.Capacity = c
	return o
}

func (o *Options[K]) SetPolicy(p policies.PolicyType) *Options[K] {
	o.Policy = p
	return o
}

func (o *Options[K]) SetNumShards(n int) *Options[K] {
	o.NumShards = n
	return o
}

func (o *Options[K]) SetHasher(h *Hasher[K]) *Options[K] {
	o.Hasher = h
	return o
}
