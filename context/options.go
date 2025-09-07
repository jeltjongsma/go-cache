package context

import (
	"errors"
	"go-cache/policies"
)

type Options[K comparable] struct {
	Capacity  int
	Policy    policies.Policy[K]
	NumShards uint64
	Hasher    *Hasher[K]
}

func NewOptions[K comparable](cap int, policy policies.Policy[K], nShards uint64) (*Options[K], error) {
	if cap < 0 {
		return nil, errors.New("capacity must be positive")
	}
	if nShards == 0 || nShards&(nShards-1) != 0 {
		return nil, errors.New("number of shards must be exponent of 2")
	}
	return &Options[K]{
		Capacity:  cap,
		Policy:    policy,
		NumShards: nShards,
	}, nil
}

func (opts *Options[K]) Equals(o *Options[K]) bool {
	optsPtype, optsKtype := opts.Policy.Type()
	oPtype, oKtype := opts.Policy.Type()
	return opts.Capacity == o.Capacity &&
		optsPtype == oPtype &&
		optsKtype == oKtype &&
		opts.NumShards == o.NumShards
}
