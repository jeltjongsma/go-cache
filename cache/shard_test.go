package cache

import (
	"go-cache/policies"
	"testing"
)

func TestShard_InitShard(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 2)
	if ptype, _ := s.policy.Type(); ptype != policies.TypeFIFO {
		t.Errorf("expected type=FIFO, got %v", ptype)
	}
	if s.cap != 2 {
		t.Errorf("expected cap=2, got %d", s.cap)
	}
}

func TestShard_Set_NoEvict(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 2)
	ok := s.Set(1, "one")
	if !ok {
		t.Fatalf("expected success, got %v", ok)
	}

	// check effects
	if len(s.store) != 1 {
		t.Fatalf("expected len=1, got %d", len(s.store))
	}
	ret, ok := s.store[1]
	if !ok {
		t.Fatalf("expected key=1 found, got false")
	}
	if ret != "one" {
		t.Errorf("expected 'one', got %s", ret)
	}
}

func TestShard_Set_Evict(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 1)

	s.Set(1, "one")
	ok := s.Set(2, "two")
	if !ok {
		t.Fatalf("expected success, got false")
	}

	// check effects
	if len(s.store) != 1 {
		t.Fatalf("expected len=1, got %d", len(s.store))
	}
	ret, ok := s.store[2]
	if !ok {
		t.Fatalf("expected key=2 found, got false")
	}
	if ret != "two" {
		t.Errorf("expected 'two', got %s", ret)
	}
	_, ok = s.store[1]
	if ok {
		t.Errorf("expected key=1 not found, got true")
	}
}

func TestShard_Set_AttemptEvictNoVictim(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 1)

	s.Set(1, "one")

	// corrupt the policy to have no keys
	s.policy = policies.NewFIFO[int]() // new empty policy

	ok := s.Set(2, "two")
	if ok {
		t.Fatalf("expected failure, got true")
	}
	// check effects
	if len(s.store) != 1 {
		t.Fatalf("expected len=1, got %d", len(s.store))
	}
	ret, ok := s.store[1]
	if !ok {
		t.Fatalf("expected key=1 found, got false")
	}
	if ret != "one" {
		t.Errorf("expected 'one', got %s", ret)
	}
	_, ok = s.store[2]
	if ok {
		t.Errorf("expected key=2 not found, got true")
	}
}

func TestShard_Set_OutOfSyncPolicy(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 1)

	// corrupt the policy to have a key not in store
	s.policy.OnSet(1) // policy has key=1, store is empty

	s.Set(2, "two")
	ok := s.Set(3, "three")
	if !ok {
		t.Fatalf("expected success, got false")
	}
	// check effects
	if len(s.store) != 1 {
		t.Fatalf("expected len=1, got %d", len(s.store))
	}
	ret, ok := s.store[3]
	if !ok {
		t.Fatalf("expected key=3 found, got false")
	}
	if ret != "three" {
		t.Errorf("expected 'three', got %s", ret)
	}
	_, ok = s.store[2]
	if ok {
		t.Errorf("expected key=2 not found, got true")
	}
	_, ok = s.store[1]
	if ok {
		t.Errorf("expected key=1 not found, got true")
	}
}

func TestShard_Get_Hit(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 2)

	s.Set(1, "one")
	ret, ok := s.Get(1)
	if !ok {
		t.Fatalf("expected hit, got miss")
	}
	if ret != "one" {
		t.Errorf("expected 'one', got %s", ret)
	}
}

func TestShard_Get_Miss(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 2)

	ret, ok := s.Get(1)
	if ok {
		t.Fatalf("expected miss, got hit")
	}
	var zero string
	if ret != zero {
		t.Errorf("expected zero value, got %s", ret)
	}
}

func TestShard_Del(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 2)

	s.Set(1, "one")
	ok := s.Del(1)
	if !ok {
		t.Fatalf("expected success, got false")
	}
	// check effects
	if len(s.store) != 0 {
		t.Fatalf("expected len=0, got %d", len(s.store))
	}
	_, ok = s.store[1]
	if ok {
		t.Errorf("expected key=1 not found, got true")
	}
}

func TestShard_Del_NotFound(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 2)

	ok := s.Del(1)
	if ok {
		t.Fatalf("expected failure, got true")
	}
	// check effects
	if len(s.store) != 0 {
		t.Fatalf("expected len=0, got %d", len(s.store))
	}
	_, ok = s.store[1]
	if ok {
		t.Errorf("expected key=1 not found, got true")
	}
}
