package cache

import (
	"testing"
	"time"

	"github.com/jeltjongsma/go-cache/pkg/policies"
)

func TestJanitor_Start(t *testing.T) {
	s := InitShard[int, int](
		policies.NewFIFO[int](),
		100,
		5*time.Minute,
	)
	j := StartJanitor(s, 0)
	if !j.shard.Equals(s) {
		t.Errorf("expected true, got false")
	}
	s.Set(1, 1)
	if !j.shard.Equals(s) {
		t.Errorf("expected true, got false")
	}
}

func TestJanitor_Stop(t *testing.T) {
	s := InitShard[int, int](
		policies.NewFIFO[int](),
		100,
		5*time.Minute,
	)
	j := StartJanitor(s, 0)
	n := time.Now()
	now := func() time.Time { return n }
	s.setNow(now)

	s.SetWithTTL(1, 1, -50)
	s.SetWithTTL(2, 2, -50)
	s.SetWithTTL(3, 3, time.Minute)

	expired := j.Stop()

	if expired != 2 {
		t.Errorf("expected 2, got %d", expired)
	}
	if e, ok := s.expiry.Peek(); !ok {
		t.Errorf("expected true, got false")
	} else if e.K != 3 {
		t.Errorf("expected k=3, got %d", e.K)
	}
}
