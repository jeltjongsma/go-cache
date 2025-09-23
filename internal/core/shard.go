package core

import (
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/jeltjongsma/go-cache/pkg/policies"
	"github.com/jeltjongsma/go-cache/pkg/ttl_queue"
)

type Entry[V any] struct {
	val       V
	expiresAt time.Time
}

type Shard[K comparable, V any] struct {
	mu         sync.RWMutex
	Store      map[K]Entry[V]
	Policy     policies.Policy[K]
	cap        int
	expiry     *ttl_queue.TTLQueue[K]
	defaultTTL time.Duration
	now        func() time.Time
}

func InitShard[K comparable, V any](Policy policies.Policy[K], cap int, defaultTTL time.Duration) *Shard[K, V] {
	return &Shard[K, V]{
		Store:      make(map[K]Entry[V], cap),
		Policy:     Policy,
		cap:        cap,
		expiry:     ttl_queue.NewTTLQueue[K](defaultTTL),
		defaultTTL: defaultTTL,
		now:        time.Now,
	}
}

func (s *Shard[K, V]) SetWithTTL(key K, val V, ttl time.Duration) (success bool, evicted int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := Entry[V]{
		val:       val,
		expiresAt: s.now().Add(ttl),
	}
	s.expiry.PushWithTTL(key, ttl)
	return s.set(key, entry)
}

func (s *Shard[K, V]) Set(key K, val V) (success bool, evicted int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := Entry[V]{
		val:       val,
		expiresAt: s.now().Add(s.defaultTTL),
	}
	s.expiry.PushStd(key)
	return s.set(key, entry)
}

func (s *Shard[K, V]) set(key K, entry Entry[V]) (success bool, evicted int) {
	_, exists := s.Store[key]

	if !exists && s.cap > 0 {
		attempts := 0
		for len(s.Store) >= s.cap {
			victim, ok := s.Policy.Evict()
			if !ok {
				return false, evicted
			}
			if _, present := s.Store[victim]; present {
				delete(s.Store, victim)
				evicted++ // only increases when evicted from Store (not Policy)
			} else {
				attempts++
				if attempts > s.cap {
					return false, evicted
				}
			}
		}
	}

	s.Store[key] = entry

	if exists {
		s.Policy.OnHit(key)
	} else {
		s.Policy.OnSet(key)
	}
	return true, evicted
}

func (s *Shard[K, V]) Get(key K) (V, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.Store[key]
	if !ok {
		var zero V
		return zero, false
	}

	now := s.now()
	if s.defaultTTL != 0 && (entry.expiresAt.Before(now) || entry.expiresAt.Equal(now)) {
		delete(s.Store, key)
		s.Policy.OnDel(key)
		s.expiry.Remove(key)
		var zero V
		return zero, false
	}

	if s.defaultTTL != 0 {
		// only ran on hits
		budget := 4
		for i := 0; i < budget; i++ {
			if !s.expiry.HasExpired() {
				break
			}
			victim := s.expiry.PopMin().K
			if _, ok := s.Store[victim]; ok {
				delete(s.Store, victim)
				s.Policy.OnDel(victim)
			}
		}
	}

	s.Policy.OnHit(key)
	return entry.val, true
}

// Peek is like get but won't affect eviction Policy.
func (s *Shard[K, V]) Peek(key K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.Store[key]
	if !ok {
		var zero V
		return zero, false
	}
	return entry.val, true
}

func (s *Shard[K, V]) Del(key K) (success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.Store[key]; !ok {
		return false
	}
	delete(s.Store, key)
	s.Policy.OnDel(key)
	return true
}

func (s *Shard[K, V]) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()

	clear(s.Store)
	s.Policy.Reset()
	s.expiry.Reset()
}

func (s *Shard[K, V]) Equals(o *Shard[K, V]) bool {
	sPtype, sKtype := s.Policy.Type()
	oPtype, oKtype := o.Policy.Type()
	return reflect.DeepEqual(s.Store, o.Store) &&
		s.cap == o.cap &&
		sPtype == oPtype &&
		sKtype == oKtype
}

func (s *Shard[K, V]) setNow(now func() time.Time) {
	s.now = now
	s.expiry.SetNow(now)
}

func (s *Shard[K, V]) Validate() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.expiry.Len() < len(s.Store) {
		return errors.New("expiry out of sync")
	}
	if len(s.Store) != s.Policy.Len() {
		return errors.New("policy out of sync")
	}
	return nil
}
