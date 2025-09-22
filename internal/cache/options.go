package cache

import (
	"go-cache/internal/policies"
	"go-cache/pkg/hasher"
	"time"
)

type Options[K comparable] struct {
	Capacity   int
	Policy     policies.PolicyType
	NumShards  int
	Hasher     *hasher.Hasher[K]
	DefaultTTL time.Duration
}

func NewOptions[K comparable]() *Options[K] {
	return &Options[K]{
		Capacity:   100,
		Policy:     policies.TypeFIFO,
		NumShards:  16,
		Hasher:     hasher.NewHasher[K](nil),
		DefaultTTL: 5 * time.Minute,
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

func (o *Options[K]) SetHasher(h *hasher.Hasher[K]) *Options[K] {
	o.Hasher = h
	return o
}

func (o *Options[K]) SetDefaultTTL(ttl time.Duration) *Options[K] {
	o.DefaultTTL = ttl
	return o
}
