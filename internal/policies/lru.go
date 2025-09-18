package policies

import (
	"fmt"
	"reflect"
)

type Node[K comparable] struct {
	key        K
	prev, next *Node[K]
}

type LRU[K comparable] struct {
	nodes      map[K]*Node[K]
	Head, Tail *Node[K]
}

func NewLRU[K comparable]() *LRU[K] {
	return &LRU[K]{nodes: make(map[K]*Node[K])}
}

func (p *LRU[K]) Type() (PolicyType, reflect.Type) {
	t := reflect.TypeOf((*K)(nil)).Elem()
	return TypeLRU, t
}

// fails silently when key is not in policy
func (p *LRU[K]) OnHit(k K) {
	if n := p.nodes[k]; n != nil && n != p.Head {
		p.detach(n)
		p.insertFront(n)
	}
}

func (p *LRU[K]) OnSet(k K) {
	if n := p.nodes[k]; n != nil {
		if n != p.Head {
			p.detach(n)
			p.insertFront(n)
		}
		return
	}
	n := &Node[K]{key: k}
	p.nodes[k] = n
	p.insertFront(n)
}

func (p *LRU[K]) OnDel(k K) {
	if n := p.nodes[k]; n != nil {
		p.detach(n)
		delete(p.nodes, k)
	}
}

func (p *LRU[K]) Evict() (K, bool) {
	if p.Tail == nil {
		var zero K
		return zero, false
	}
	n := p.Tail
	p.detach(n)
	delete(p.nodes, n.key)
	return n.key, true
}

func (p *LRU[K]) Reset() {
	clear(p.nodes)
	p.Head, p.Tail = nil, nil
}

func (p *LRU[K]) detach(n *Node[K]) {
	switch {
	case n.prev == nil && n.next == nil:
		// only node
		p.Head, p.Tail = nil, nil
	case n.prev == nil:
		// n is head
		p.Head = n.next
		p.Head.prev = nil
	case n.next == nil:
		// n is tail
		p.Tail = n.prev
		p.Tail.next = nil
	default:
		// middle
		n.prev.next = n.next
		n.next.prev = n.prev
	}
	n.prev, n.next = nil, nil
}

func (p *LRU[K]) insertFront(n *Node[K]) {
	if p.Head == nil {
		p.Head, p.Tail = n, n
		return
	}
	n.next = p.Head
	p.Head.prev = n
	p.Head = n
}

func (p *LRU[K]) Validate() error {
	current := p.Head
	count := 0
	// check if all nodes in list are in map
	for current != nil {
		if n, ok := p.nodes[current.key]; !ok {
			return fmt.Errorf("node '%v' not in map", n)
		}
		count++
		current = current.next
	}
	// check if map and list contain same amount of nodes
	if len(p.nodes) != count {
		return fmt.Errorf("len(map): %d != len(list): %d", len(p.nodes), count)
	}
	return nil
}

func (p LRU[K]) Len() int {
	count := 0
	current := p.Head
	for current != nil {
		count++
		current = current.next
	}
	return count
}

func (p *LRU[K]) Equals(o Policy[any]) bool {
	pPtype, pKtype := p.Type()
	oPtype, oKtype := p.Type()
	return pPtype == oPtype && pKtype == oKtype
}
