package cache

import (
	"go-cache/context"
	"go-cache/policies"
	"testing"
)

func TestNew(t *testing.T) {
	opts, _ := context.NewOptions(2, policies.NewFIFO[int](), 2)
	c := New[int, string](opts)
	if ptype, _ := c.policy.Type(); ptype != policies.TypeFIFO {
		t.Errorf("expected type=FIFO, got %v", ptype)
	}

	if opts != c.opts {
		t.Errorf("options don't match")
	}
}

func TestSet_NoEvict(t *testing.T) {
	opts, _ := context.NewOptions(1, policies.NewFIFO[int](), 2)
	c := New[int, string](opts)

	ok := c.Set(1, "one")
	if !ok {
		t.Fatalf("expected success, got %v", ok)
	}

	// check effects
	if len(c.store) != 1 {
		t.Fatalf("expected len=1, got %d", len(c.store))
	}
	ret, ok := c.store[1]
	if !ok {
		t.Fatalf("expected key=1 found, got false")
	}
	if ret != "one" {
		t.Errorf("expected 'one', got %s", ret)
	}
}

func TestSet_Evict(t *testing.T) {
	opts, _ := context.NewOptions(1, policies.NewFIFO[int](), 2)
	c := New[int, string](opts)

	c.Set(1, "one")
	ok := c.Set(2, "two")
	if !ok {
		t.Fatalf("expected success, got false")
	}

	// check effects
	if len(c.store) != 1 {
		t.Fatalf("expected len=1, got %d", len(c.store))
	}
	ret, ok := c.store[2]
	if !ok {
		t.Fatalf("expected key=2 found, got false")
	}
	if ret != "two" {
		t.Errorf("expected 'two', got %s", ret)
	}
	_, ok = c.store[1]
	if ok {
		t.Errorf("expected key=1 not found, got true")
	}
}

func TestSet_AttemptEvictNoVictim(t *testing.T) {
	opts, _ := context.NewOptions(1, policies.NewFIFO[int](), 2)
	c := New[int, string](opts)

	c.Set(1, "one")

	// corrupt the policy to have no keys
	c.policy = policies.NewFIFO[int]() // new empty policy

	ok := c.Set(2, "two")
	if ok {
		t.Fatalf("expected failure, got true")
	}
	// check effects
	if len(c.store) != 1 {
		t.Fatalf("expected len=1, got %d", len(c.store))
	}
	ret, ok := c.store[1]
	if !ok {
		t.Fatalf("expected key=1 found, got false")
	}
	if ret != "one" {
		t.Errorf("expected 'one', got %s", ret)
	}
	_, ok = c.store[2]
	if ok {
		t.Errorf("expected key=2 not found, got true")
	}
}

func TestSet_OutOfSyncPolicy(t *testing.T) {
	opts, _ := context.NewOptions(1, policies.NewFIFO[int](), 2)
	c := New[int, string](opts)

	// corrupt the policy to have a key not in store
	c.policy.OnSet(1) // policy has key=1, store is empty

	c.Set(2, "two")
	ok := c.Set(3, "three")
	if !ok {
		t.Fatalf("expected success, got false")
	}
	// check effects
	if len(c.store) != 1 {
		t.Fatalf("expected len=1, got %d", len(c.store))
	}
	ret, ok := c.store[3]
	if !ok {
		t.Fatalf("expected key=3 found, got false")
	}
	if ret != "three" {
		t.Errorf("expected 'three', got %s", ret)
	}
	_, ok = c.store[2]
	if ok {
		t.Errorf("expected key=2 not found, got true")
	}
	_, ok = c.store[1]
	if ok {
		t.Errorf("expected key=1 not found, got true")
	}
}

func TestGet_Hit(t *testing.T) {
	opts, _ := context.NewOptions(1, policies.NewFIFO[int](), 2)
	c := New[int, string](opts)

	c.Set(1, "one")
	ret, ok := c.Get(1)
	if !ok {
		t.Fatalf("expected hit, got miss")
	}
	if ret != "one" {
		t.Errorf("expected 'one', got %s", ret)
	}
}

func TestGet_Miss(t *testing.T) {
	opts, _ := context.NewOptions(1, policies.NewFIFO[int](), 2)
	c := New[int, string](opts)

	ret, ok := c.Get(1)
	if ok {
		t.Fatalf("expected miss, got hit")
	}
	var zero string
	if ret != zero {
		t.Errorf("expected zero value, got %s", ret)
	}
}

func TestDel(t *testing.T) {
	opts, _ := context.NewOptions(1, policies.NewFIFO[int](), 2)
	c := New[int, string](opts)

	c.Set(1, "one")
	ok := c.Del(1)
	if !ok {
		t.Fatalf("expected success, got false")
	}
	// check effects
	if len(c.store) != 0 {
		t.Fatalf("expected len=0, got %d", len(c.store))
	}
	_, ok = c.store[1]
	if ok {
		t.Errorf("expected key=1 not found, got true")
	}
}

func TestDel_NotFound(t *testing.T) {
	opts, _ := context.NewOptions(1, policies.NewFIFO[int](), 2)
	c := New[int, string](opts)

	ok := c.Del(1)
	if ok {
		t.Fatalf("expected failure, got true")
	}
	// check effects
	if len(c.store) != 0 {
		t.Fatalf("expected len=0, got %d", len(c.store))
	}
	_, ok = c.store[1]
	if ok {
		t.Errorf("expected key=1 not found, got true")
	}
}
