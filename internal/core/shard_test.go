package core

import (
	"testing"
	"time"

	"github.com/jeltjongsma/go-cache/pkg/policies"
)

func TestShard_InitShard(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 2, 100)
	if ptype, _ := s.Policy.Type(); ptype != policies.TypeFIFO {
		t.Errorf("expected type=FIFO, got %v", ptype)
	}
	if s.cap != 2 {
		t.Errorf("expected cap=2, got %d", s.cap)
	}
}

func TestShard_Set_TTL(t *testing.T) {
	defaultTTL := time.Duration(100)
	s := InitShard[int, int](policies.NewFIFO[int](), 100, defaultTTL)
	n := time.Now()
	s.setNow(func() time.Time { return n })

	s.Set(1, 1)
	if e := s.Store[1]; !e.expiresAt.Equal(n.Add(100)) {
		t.Errorf("expected %v + 100, got %v", n, e.expiresAt)
	}
	s.SetWithTTL(2, 2, 50)
	if e := s.Store[2]; !e.expiresAt.Equal(n.Add(50)) {
		t.Errorf("expected %v + 50, got %v", n, e.expiresAt)
	}
}

func TestShard_Set_NoEvict(t *testing.T) {
	// check for success
	s := InitShard[int, int](policies.NewFIFO[int](), 100, 100)
	for i := range 10000 {
		ok, _ := s.Set(i, i)
		if !ok {
			t.Fatalf("expected true, got false")
		}
	}

	s = InitShard[int, int](policies.NewFIFO[int](), 2, 100)
	ok, evicted := s.Set(1, 1)
	if !ok {
		t.Fatalf("expected success, got %v", ok)
	}
	if evicted != 0 {
		t.Fatalf("expected evicted=0, got %d", evicted)
	}

	// check effects
	if len(s.Store) != 1 {
		t.Fatalf("expected len=1, got %d", len(s.Store))
	}
	ret, ok := s.Store[1]
	if !ok {
		t.Fatalf("expected key=1 found, got false")
	}
	if ret.val != 1 {
		t.Errorf("expected '1', got %d", ret.val)
	}
}

func TestShard_Set_Evict(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 1, 100)

	s.Set(1, "one")
	ok, evicted := s.Set(2, "two")
	if !ok {
		t.Fatalf("expected success, got false")
	}
	if evicted != 1 {
		t.Fatalf("expected evicted=1, got %d", evicted)
	}

	// check effects
	if len(s.Store) != 1 {
		t.Fatalf("expected len=1, got %d", len(s.Store))
	}
	ret, ok := s.Store[2]
	if !ok {
		t.Fatalf("expected key=2 found, got false")
	}
	if ret.val != "two" {
		t.Errorf("expected 'two', got %s", ret.val)
	}
	_, ok = s.Store[1]
	if ok {
		t.Errorf("expected key=1 not found, got true")
	}
}

func TestShard_Set_AttemptEvictNoVictim(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 1, 100)

	s.Set(1, "one")

	// corrupt the Policy to have no keys
	s.Policy = policies.NewFIFO[int]() // new empty Policy

	ok, evicted := s.Set(2, "two")
	if ok {
		t.Fatalf("expected failure, got true")
	}
	if evicted != 0 {
		t.Fatalf("expected evicted=0, got %d", evicted)
	}
	// check effects
	if len(s.Store) != 1 {
		t.Fatalf("expected len=1, got %d", len(s.Store))
	}
	ret, ok := s.Store[1]
	if !ok {
		t.Fatalf("expected key=1 found, got false")
	}
	if ret.val != "one" {
		t.Errorf("expected 'one', got %s", ret.val)
	}
	_, ok = s.Store[2]
	if ok {
		t.Errorf("expected key=2 not found, got true")
	}
}

func TestShard_Set_OutOfSyncPolicy(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 1, 100)

	// corrupt the Policy to have a key not in Store
	s.Policy.OnSet(1) // Policy has key=1, Store is empty

	s.Set(2, "two")
	ok, evicted := s.Set(3, "three")
	if !ok {
		t.Fatalf("expected success, got false")
	}
	// evicted only counts evictions from Store (not Policy)
	if evicted != 1 {
		t.Fatalf("expected evicted=1, got %d", evicted)
	}
	// check effects
	if len(s.Store) != 1 {
		t.Fatalf("expected len=1, got %d", len(s.Store))
	}
	ret, ok := s.Store[3]
	if !ok {
		t.Fatalf("expected key=3 found, got false")
	}
	if ret.val != "three" {
		t.Errorf("expected 'three', got %s", ret.val)
	}
	_, ok = s.Store[2]
	if ok {
		t.Errorf("expected key=2 not found, got true")
	}
	_, ok = s.Store[1]
	if ok {
		t.Errorf("expected key=1 not found, got true")
	}
}

func TestShard_Get_Hit(t *testing.T) {
	s := InitShard[int, int](policies.NewFIFO[int](), 100, 100)
	n := time.Now()
	s.setNow(func() time.Time { return n })

	// fill shard
	for i := range 100 {
		s.Set(i, i)
	}

	// test if all are hit
	for i := range 100 {
		ret, ok := s.Get(i)
		if !ok {
			t.Fatalf("expected hit, got miss")
		}
		if ret != i {
			t.Errorf("expected '%d', got %d", i, ret)
		}
	}
}

func TestShard_Get_Miss(t *testing.T) {
	s := InitShard[int, int](policies.NewFIFO[int](), 100, 100)
	n := time.Now()
	s.setNow(func() time.Time { return n })

	for i := range 50 {
		s.Set(i, i)
	}

	for i := 50; i < 100; i++ {
		ret, ok := s.Get(i)
		if ok {
			t.Fatalf("expected miss, got hit")
		}
		var zero int
		if ret != zero {
			t.Errorf("expected zero value, got %d", ret)
		}
	}
}

func TestShard_Get_Expiry(t *testing.T) {
	s := InitShard[int, int](policies.NewLRU[int](), 10, 100)
	n := time.Now()
	s.setNow(func() time.Time { return n })

	s.SetWithTTL(1, 1, 0) // equal
	_, ok := s.Get(1)
	if ok {
		t.Errorf("expected false, got true")
	}
	s.SetWithTTL(2, 2, -50) // past
	_, ok = s.Get(2)
	if ok {
		t.Errorf("expected false, got true")
	}
	s.SetWithTTL(3, 3, 50) // future
	_, ok = s.Get(3)
	if !ok {
		t.Errorf("expected true, got false")
	}
}

func TestShard_Get_ExpiryBudget(t *testing.T) {
	s := InitShard[int, int](policies.NewLRU[int](), 10, 100)
	n := time.Now()
	s.setNow(func() time.Time { return n })

	for i := range 5 { // 1 more than budget
		s.SetWithTTL(i, i, -50)
	}
	s.SetWithTTL(5, 5, 50)

	s.Get(5) // hit: should run cleanup

	if !s.expiry.HasExpired() {
		t.Fatalf("expected true, got false")
	}
	if l := s.expiry.Len(); l != 2 { // 4, 5
		t.Fatalf("expected len=2, got %d", l)
	}
	e, ok := s.expiry.Peek()
	if !ok {
		t.Fatalf("expected true, got false")
	}
	if e.K != 4 {
		t.Errorf("expected k=4, got %d", e.K)
	}

	s.Flush() // reset shard

	s.SetWithTTL(0, 0, -50)
	s.SetWithTTL(1, 1, -50)

	s.Get(0) // expired: shouldn't run cleanup

	if l := s.expiry.Len(); l != 1 {
		t.Fatalf("expected len=1, got %d", l)
	}
	e, ok = s.expiry.Peek()
	if !ok {
		t.Fatalf("expected true, got false")
	}
	if e.K != 1 {
		t.Errorf("expected k=1, got %d", e.K)
	}

	s.Flush() // reset shard

	s.SetWithTTL(0, 0, -50)

	s.Get(1) // miss: shouldn't run cleanup

	if l := s.expiry.Len(); l != 1 {
		t.Fatalf("expected len=1, got %d", l)
	}
	e, ok = s.expiry.Peek()
	if !ok {
		t.Fatalf("expected true, got false")
	}
	if e.K != 0 {
		t.Errorf("expected k=0, got %d", e.K)
	}
}

func TestShard_Get_NoTTL(t *testing.T) {
	s := InitShard[int, int](policies.NewLRU[int](), 10, 0)
	n := time.Now()
	s.setNow(func() time.Time { return n })

	s.SetWithTTL(1, 1, -50)
	s.SetWithTTL(2, 2, 50)

	s.Get(2) // hit, but DefaultTTL == 0: shouldn't run cleanup

	if !s.expiry.HasExpired() {
		t.Fatalf("expected true, got false")
	}
	if l := s.expiry.Len(); l != 2 { // 1, 2
		t.Fatalf("expected len=2, got %d", l)
	}
	e, ok := s.expiry.Peek()
	if !ok {
		t.Fatalf("expected true, got false")
	}
	if e.K != 1 {
		t.Errorf("expected k=1, got %d", e.K)
	}
}

func TestShard_Peek_Hit(t *testing.T) {
	s := InitShard[int, string](policies.NewLRU[int](), 2, 100)

	s.Set(1, "one")
	s.Set(2, "two")
	ret, ok := s.Peek(1)
	if !ok {
		t.Fatalf("expected hit, got miss")
	}
	if ret != "one" {
		t.Errorf("expected 'one', got %s", ret)
	}

	// peek shouldn't have side effects on policy, so '1' should still be victim
	victim, ok := s.Policy.Evict()
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	if victim != 1 {
		t.Errorf("expected victim=1, got %d", victim)
	}
}

func TestShard_Peek_Miss(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 2, 100)

	ret, ok := s.Peek(1)
	if ok {
		t.Fatalf("expected miss, got hit")
	}
	var zero string
	if ret != zero {
		t.Errorf("expected zero value, got %s", ret)
	}
}

func TestShard_Del(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 2, 100)

	s.Set(1, "one")
	ok := s.Del(1)
	if !ok {
		t.Fatalf("expected success, got false")
	}
	// check effects
	if len(s.Store) != 0 {
		t.Fatalf("expected len=0, got %d", len(s.Store))
	}
	_, ok = s.Store[1]
	if ok {
		t.Errorf("expected key=1 not found, got true")
	}
}

func TestShard_Del_NotFound(t *testing.T) {
	s := InitShard[int, string](policies.NewFIFO[int](), 2, 100)

	ok := s.Del(1)
	if ok {
		t.Fatalf("expected failure, got true")
	}
	// check effects
	if len(s.Store) != 0 {
		t.Fatalf("expected len=0, got %d", len(s.Store))
	}
	_, ok = s.Store[1]
	if ok {
		t.Errorf("expected key=1 not found, got true")
	}
}

func TestShard_Flush(t *testing.T) {
	s := InitShard[int, int](policies.NewLRU[int](), 2, 100)

	s.Set(1, 1)
	s.Set(2, 2)
	s.Set(3, 3)
	s.Set(4, 4)
	s.Set(5, 5)

	s.Flush()

	// check effects
	if len(s.Store) != 0 {
		t.Fatalf("expected Store.len=0, got %d", len(s.Store))
	}
	victim, ok := s.Policy.Evict()
	if ok {
		t.Fatalf("expected ok=false, got true")
	}
	var zero int
	if victim != zero {
		t.Fatalf("expected zero value, got %v", victim)
	}

	if _, ok := s.expiry.Peek(); ok {
		t.Fatalf("expected false, got true")
	}
}

func TestShard_Flush_AlreadyEmtpy(t *testing.T) {
	s := InitShard[int, int](policies.NewLRU[int](), 2, 100)

	s.Flush()

	// check effects
	if len(s.Store) != 0 {
		t.Fatalf("expected Store.len=0, got %d", len(s.Store))
	}
	victim, ok := s.Policy.Evict()
	if ok {
		t.Fatalf("expected ok=false, got true")
	}
	var zero int
	if victim != zero {
		t.Fatalf("expected zero value, got %v", victim)
	}
}

func TestShard_setNow(t *testing.T) {
	s := InitShard[int, int](policies.NewLRU[int](), 2, 100)
	n := time.Now()
	s.setNow(func() time.Time { return n })

	if !s.now().Equal(n) {
		t.Errorf("now not set")
	}
}
