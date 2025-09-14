package policies

import (
	"testing"
)

func TestNewFIFO(t *testing.T) {
	p := NewFIFO[string]()
	if p.keys == nil {
		t.Fatal("Expected 'keys' to be set, got nil")
	}
	if len(p.keys) != 0 {
		t.Errorf("Expected 0, got %d", len(p.keys))
	}
}

func TestFIFO_Type(t *testing.T) {
	p := NewFIFO[int]()
	ptype, ktype := p.Type()
	if ptype != TypeFIFO {
		t.Errorf("expected 'FIFO', got %v", ptype)
	}
	if ktype.String() != "int" {
		t.Errorf("expected 'int', got %s", ktype.String())
	}
}

// no side effects
func TestFIFO_OnHit(t *testing.T) {
	p := NewFIFO[string]()
	p.OnHit("")
	if len(p.keys) != 0 {
		t.Errorf("Expected 0, got %d", len(p.keys))
	}
}

func TestFIFO_OnSet(t *testing.T) {
	p := NewFIFO[string]()
	p.OnSet("one")
	if len(p.keys) != 1 {
		t.Fatalf("Expected 1, got %d", len(p.keys))
	}
	if p.keys[0] != "one" {
		t.Errorf("Expected 'one', got %s", p.keys[0])
	}
	p.OnSet("two")
	if len(p.keys) != 2 {
		t.Fatalf("Expected 1, got %d", len(p.keys))
	}
	if p.keys[1] != "two" {
		t.Errorf("Expected 'two', got %s", p.keys[1])
	}
}

func TestFIFO_OnSet_Repeat(t *testing.T) {
	p := NewFIFO[string]()
	p.OnSet("one")
	if len(p.keys) != 1 {
		t.Fatalf("Expected 1, got %d", len(p.keys))
	}
	if p.keys[0] != "one" {
		t.Errorf("Expected 'one', got %s", p.keys[0])
	}
	p.OnSet("one")
	if len(p.keys) != 2 {
		t.Fatalf("Expected 1, got %d", len(p.keys))
	}
	if p.keys[1] != "one" {
		t.Errorf("Expected 'one', got %s", p.keys[1])
	}
}

func TestFIFO_OnDel(t *testing.T) {
	p := NewFIFO[string]()
	p.OnSet("one")
	p.OnDel("one")
	if len(p.keys) != 0 {
		t.Fatalf("Expected 0, got %d", len(p.keys))
	}
}

func TestFIFO_OnDel_MultipleValues(t *testing.T) {
	p := NewFIFO[int]()
	p.OnSet(1)
	p.OnSet(2)
	p.OnDel(2)
	if len(p.keys) != 1 {
		t.Fatalf("Expected 1, got %d", len(p.keys))
	}
	if p.keys[0] != 1 {
		t.Errorf("expected 1, got %d", p.keys[0])
	}
}

func TestFIFO_Evict(t *testing.T) {
	p := NewFIFO[int]()
	p.OnSet(1)
	p.OnSet(2)
	k, ok := p.Evict()
	if !ok {
		t.Fatalf("expected evict=true, got %v", ok)
	}
	if k != 1 {
		t.Errorf("expected k=1, got %d", k)
	}
	if len(p.keys) != 1 {
		t.Fatalf("expected len=1, got %d", len(p.keys))
	}
	if p.keys[0] != 2 {
		t.Errorf("expected [0]=2, got %d", p.keys[0])
	}
}

func TestFIFO_Evict_Empty(t *testing.T) {
	p := NewFIFO[int]()
	_, ok := p.Evict()
	if ok {
		t.Errorf("expected evict=false, got %v", ok)
	}
}
