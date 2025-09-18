package policies

import "testing"

func TestNewLRU(t *testing.T) {
	p := NewLRU[int]()
	if p.nodes == nil {
		t.Fatalf("expected nodes to be set, got nil")
	}
	if len(p.nodes) != 0 {
		t.Errorf("expected len=0, got %d", len(p.nodes))
	}
	if p.Head != nil || p.Tail != nil {
		t.Errorf("expected head & tail = nil, got %v, %v", p.Head, p.Tail)
	}
}

func TestLRU_Type(t *testing.T) {
	p := NewLRU[int]()
	ptype, ktype := p.Type()
	if ptype != TypeLRU {
		t.Errorf("expected 'LRU', got %v", ptype)
	}
	if ktype.String() != "int" {
		t.Errorf("expected 'int', got %s", ktype.String())
	}
}

func TestLRU_OnHit(t *testing.T) {
	p := NewLRU[int]()
	p.OnSet(1)
	p.OnSet(2)

	p.OnHit(1)
	if p.Head.key != 1 {
		t.Errorf("expected '1' in front, got %d", p.Head.key)
	}
	if len(p.nodes) != 2 {
		t.Errorf("expected len=2, got %d", len(p.nodes))
	}

	if err := p.Validate(); err != nil {
		t.Fatalf("policy not valid: %v", err)
	}

	// key not in policy shouldn't affect policy
	p.OnHit(3)
	if p.Head.key != 1 {
		t.Errorf("expected '1' in front, got %d", p.Head.key)
	}
	if len(p.nodes) != 2 {
		t.Errorf("expected len=1, got %d", len(p.nodes))
	}

	if err := p.Validate(); err != nil {
		t.Fatalf("policy not valid: %v", err)
	}
}

func TestLRU_OnSet(t *testing.T) {
	p := NewLRU[int]()

	for i := 0; i < 10; i++ {
		p.OnSet(i)
	}

	if err := p.Validate(); err != nil {
		t.Fatalf("policy not valid: %v", err)
	}

	if len(p.nodes) != 10 {
		t.Fatalf("expected len=10, got %d", len(p.nodes))
	}

	current := p.Head
	key := 9
	for current.next != nil {
		if current.key != key {
			t.Errorf("expected key=%d, got %d", key, current.key)
		}
		key--
		current = current.next
	}
}

func TestLRU_OnSet_Duplicate(t *testing.T) {
	p := NewLRU[int]()
	p.OnSet(1)
	p.OnSet(2)
	p.OnSet(1)

	if err := p.Validate(); err != nil {
		t.Fatalf("policy not valid: %v", err)
	}

	if len(p.nodes) != 2 {
		t.Fatalf("expected len=2, got %d", len(p.nodes))
	}

	if p.Head.key != 1 {
		t.Errorf("expected key=1, got %d", p.Head.key)
	}
}

func TestLRU_OnDel(t *testing.T) {
	p := NewLRU[int]()
	p.OnSet(1)
	p.OnSet(2)

	p.OnDel(1)

	if err := p.Validate(); err != nil {
		t.Fatalf("policy not valid: %v", err)
	}

	if l := len(p.nodes); l != 1 {
		t.Fatalf("expected len=1, got %d", l)
	}

	if k := p.Head.key; k != 2 {
		t.Errorf("expected key=2, got %d", k)
	}
}

func TestLRU_detach(t *testing.T) {
	p := NewLRU[int]()
	p.OnSet(1)
	p.OnSet(2)
	p.OnSet(3)
	p.OnSet(4)

	// detach middle
	p.detach(p.nodes[3])
	if p.Len() != 3 {
		t.Fatalf("expected len=3, got %d", len(p.nodes))
	}
	if p.Head.key != 4 {
		t.Errorf("expected head=4, got %d", p.Head.key)
	}
	if p.Tail.key != 1 {
		t.Errorf("expected tail=1, got %d", p.Tail.key)
	}
	// detach tail
	p.detach(p.Tail)
	if p.Len() != 2 {
		t.Fatalf("expected len=2, got %d", len(p.nodes))
	}
	if p.Head.key != 4 {
		t.Errorf("expected head=4, got %d", p.Head.key)
	}
	if p.Tail.key != 2 {
		t.Errorf("expected tail=2, got %d", p.Tail.key)
	}
	// detach head
	p.detach(p.Head)
	if p.Len() != 1 {
		t.Fatalf("expected len=1, got %d", len(p.nodes))
	}
	if p.Head.key != 2 {
		t.Errorf("expected head=2, got %d", p.Head.key)
	}
	if p.Tail.key != 2 {
		t.Errorf("expected tail=2, got %d", p.Tail.key)
	}
	// detach only node
	p.detach(p.Head)
	if p.Len() != 0 {
		t.Fatalf("expected len=0, got %d", len(p.nodes))
	}
	if p.Head != nil {
		t.Errorf("expected head=nil, got %v", p.Head)
	}
	if p.Tail != nil {
		t.Errorf("expected tail=nil, got %v", p.Tail)
	}
}

func TestLRU_insertFront(t *testing.T) {
	p := NewLRU[int]()
	n1 := &Node[int]{key: 1}
	p.insertFront(n1)
	if p.Head.key != 1 || p.Tail.key != 1 {
		t.Errorf("expected head & tail = 1, got %v, %v", p.Head, p.Tail)
	}
	n2 := &Node[int]{key: 2}
	p.insertFront(n2)
	if p.Head.key != 2 {
		t.Errorf("expected head = 2, got %v", p.Head)
	}
	if p.Tail.key != 1 {
		t.Errorf("expected tail = 1, got %v", p.Tail)
	}
	n3 := &Node[int]{key: 3}
	p.insertFront(n3)
	if p.Head.key != 3 {
		t.Errorf("expected head = 3, got %v", p.Head)
	}
	if p.Tail.key != 1 {
		t.Errorf("expected tail = 1, got %v", p.Tail)
	}
	if p.Len() != 3 {
		t.Errorf("expected len=3, got %d", len(p.nodes))
	}
}

func TestLRU_Validate(t *testing.T) {
	p := NewLRU[int]()
	p.OnSet(1)
	p.OnSet(2)
	p.OnSet(3)
	if err := p.Validate(); err != nil {
		t.Fatalf("policy not valid: %v", err)
	}
	p.nodes[4] = &Node[int]{key: 4} // add node not in list
	if err := p.Validate(); err == nil {
		t.Fatalf("expected error, got nil")
	}
	// reset
	p = NewLRU[int]()
	p.OnSet(1)
	p.OnSet(2)
	p.OnSet(3)
	p.OnSet(4)

	delete(p.nodes, 2) // remove node in list
	if err := p.Validate(); err == nil {
		t.Fatalf("expected error, got nil")
	}

	// reset
	p = NewLRU[int]()
	p.OnSet(1)
	p.OnSet(2)
	p.OnSet(3)
	p.OnSet(4)

	p.Tail.next = &Node[int]{key: 5} // corrupt list
	if err := p.Validate(); err == nil {
		t.Fatalf("expected error, got nil")
	}
	p.Tail.next = nil
	if err := p.Validate(); err != nil {
		t.Fatalf("policy not valid: %v", err)
	}
	// remove all nodes
	p.OnDel(1)
	p.OnDel(2)
	p.OnDel(3)
	p.OnDel(4)

	if err := p.Validate(); err != nil {
		t.Fatalf("policy not valid: %v", err)
	}
	if p.Len() != 0 {
		t.Errorf("expected len=0, got %d", len(p.nodes))
	}
	if p.Head != nil {
		t.Errorf("expected head=nil, got %v", p.Head)
	}
	if p.Tail != nil {
		t.Errorf("expected tail=nil, got %v", p.Tail)
	}
	if len(p.nodes) != 0 {
		t.Errorf("expected len=0, got %d", len(p.nodes))
	}
}

func TestLRU_Len(t *testing.T) {
	p := NewLRU[int]()
	p.OnSet(1)
	p.OnSet(2)
	p.OnSet(3)

	if p.Len() != 3 {
		t.Errorf("expected 3, got %d", p.Len())
	}
}
