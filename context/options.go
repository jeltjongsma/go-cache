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

func (o *Options[K]) Equals(other *Options[K]) bool {
	return o.Capacity == other.Capacity &&
		o.Policy.Type() == other.Policy.Type()
}
