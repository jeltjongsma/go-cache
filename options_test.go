package cache

import (
	"testing"
	"time"

	"github.com/jeltjongsma/go-cache/pkg/policies"
)

func TestOptions_Capacity(t *testing.T) {
	opts := NewOptions[int]()
	if opts.Capacity != 1000 {
		t.Errorf("expected 1000, got %d", opts.Capacity)
	}
	opts.SetCapacity(10)
	if opts.Capacity != 10 {
		t.Errorf("expected 10, got %d", opts.Capacity)
	}
}

func TestOptions_Policy(t *testing.T) {
	opts := NewOptions[int]()
	if opts.Policy != policies.TypeFIFO {
		t.Errorf("expected fifo, got %s", opts.Policy)
	}
	opts.SetPolicy(policies.TypeLRU)
	if opts.Policy != policies.TypeLRU {
		t.Errorf("expected lru, got %s", opts.Policy)
	}
}

func TestOptions_NumShards(t *testing.T) {
	opts := NewOptions[int]()
	if opts.NumShards != 16 {
		t.Errorf("expected 16, got %d", opts.NumShards)
	}
	opts.SetNumShards(8)
	if opts.NumShards != 8 {
		t.Errorf("expected 8, got %d", opts.NumShards)
	}
}

func TestOptions_Hasher(t *testing.T) {
	opts := NewOptions[int]()
	if opts.Hasher == nil {
		t.Errorf("expected hasher, got nil")
	}
}

func TestOptions_DefaultTTL(t *testing.T) {
	opts := NewOptions[int]()
	if opts.DefaultTTL != 5*time.Minute {
		t.Errorf("expected 5 minutes, got %v", opts.DefaultTTL)
	}
	opts.SetDefaultTTL(10)
	if opts.DefaultTTL != 10 {
		t.Errorf("expected 10, got %v", opts.DefaultTTL)
	}
}
