package context

import (
	"go-cache/policies"
)

type Options[K comparable] struct {
	Capacity  int
	Policy    policies.PolicyType
	NumShards int
	Hasher    *Hasher[K]
}
