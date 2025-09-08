package policies

import "reflect"

type FIFO[K comparable] struct {
	keys []K
}

func NewFIFO[K comparable]() *FIFO[K] {
	return &FIFO[K]{
		keys: make([]K, 0),
	}
}

func (p *FIFO[K]) OnHit(key K) {
	// No action needed on hit for FIFO
}

func (p *FIFO[K]) OnSet(key K) {
	p.keys = append(p.keys, key)
}

func (p *FIFO[K]) OnDel(key K) {
	for i, k := range p.keys {
		if k == key {
			p.keys = append(p.keys[:i], p.keys[i+1:]...)
			break
		}
	}
}

func (p *FIFO[K]) Evict() (K, bool) {
	if len(p.keys) == 0 {
		return *new(K), false
	}
	evictedKey := p.keys[0]
	p.keys = p.keys[1:]
	return evictedKey, true
}

func (p *FIFO[K]) Type() (PolicyType, reflect.Type) {
	t := reflect.TypeOf((*K)(nil)).Elem()
	return TypeFIFO, t
}

func (p *FIFO[K]) Equals(o Policy[any]) bool {
	pPtype, pKtype := p.Type()
	oPtype, oKtype := p.Type()
	return pPtype == oPtype && pKtype == oKtype
}
