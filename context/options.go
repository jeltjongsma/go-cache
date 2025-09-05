package context

import (
	"errors"
	"go-cache/policies"
)

type Options[K comparable] struct {
	Capacity int
	Policy   policies.Policy[K]
}

func NewOptions[K comparable](cap int, policy policies.Policy[K]) (*Options[K], error) {
	if cap < 0 {
		return nil, errors.New("capacity must be positive")
	}
	return &Options[K]{
		Capacity: cap,
		Policy:   policy,
	}, nil
}

func (opts *Options[K]) Equals(o *Options[K]) bool {
	optsPtype, optsKtype := opts.Policy.Type()
	oPtype, oKtype := opts.Policy.Type()
	return opts.Capacity == o.Capacity &&
		optsPtype == oPtype &&
		optsKtype == oKtype
}
