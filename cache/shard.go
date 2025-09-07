package cache

import (
	"go-cache/policies"
	"reflect"
	"sync"
)

type Shard[K comparable, V any] struct {
	mu     sync.RWMutex
	store  map[K]V
	policy policies.Policy[K]
	cap    int
}

func InitShard[K comparable, V any](policy policies.Policy[K], cap int) *Shard[K, V] {
	return &Shard[K, V]{
		store:  make(map[K]V, cap),
		policy: policy,
		cap:    cap,
	}
}

func (s *Shard[K, V]) Set(key K, val V) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.store[key]

	if !exists && s.cap > 0 {
		attempts := 0
		for len(s.store) >= s.cap {
			victim, ok := s.policy.Evict()
			if !ok {
				return false
			}
			if _, present := s.store[victim]; present {
				delete(s.store, victim)
			} else {
				attempts++
				if attempts > s.cap {
					return false
				}
			}
		}
	}

	s.store[key] = val
	s.policy.OnSet(key)
	return true
}

func (s *Shard[K, V]) Get(key K) (V, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.store[key]
	if !ok {
		var zero V
		return zero, false
	}
	s.policy.OnHit(key)
	return val, true
}

func (s *Shard[K, V]) Del(key K) (success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.store[key]; !ok {
		return false
	}
	delete(s.store, key)
	s.policy.OnDel(key)
	return true
}

func (s *Shard[K, V]) Equals(o *Shard[K, V]) bool {
	sPtype, sKtype := s.policy.Type()
	oPtype, oKtype := o.policy.Type()
	return reflect.DeepEqual(s.store, o.store) &&
		s.cap == o.cap &&
		sPtype == oPtype &&
		sKtype == oKtype
}
