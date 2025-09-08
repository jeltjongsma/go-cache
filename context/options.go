package context

import (
	"go-cache/policies"
)

type Options[K comparable] struct {
	Capacity  int
	Policy    policies.Policy[K]
	NumShards uint64
	Hasher    *Hasher[K]
}
