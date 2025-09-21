package cache

import (
	"errors"
	"go-cache/internal/policies"
	"go-cache/pkg/ttl_queue"
	"reflect"
	"sync"
	"time"
)

type Entry[V any] struct {
	val       V
	expiresAt time.Time
}

type Shard[K comparable, V any] struct {
	mu         sync.RWMutex
	store      map[K]Entry[V]
	policy     policies.Policy[K]
	cap        int
	expiry     *ttl_queue.TTLQueue[K]
	defaultTTL time.Duration
	now        func() time.Time
}

func InitShard[K comparable, V any](policy policies.Policy[K], cap int, defaultTTL time.Duration) *Shard[K, V] {
	return &Shard[K, V]{
		store:      make(map[K]Entry[V], cap),
		policy:     policy,
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
	_, exists := s.store[key]

	if !exists && s.cap > 0 {
		attempts := 0
		for len(s.store) >= s.cap {
			victim, ok := s.policy.Evict()
			if !ok {
				return false, evicted
			}
			if _, present := s.store[victim]; present {
				delete(s.store, victim)
				evicted++ // only increases when evicted from store (not policy)
			} else {
				attempts++
				if attempts > s.cap {
					return false, evicted
				}
			}
		}
	}

	s.store[key] = entry

	if exists {
		s.policy.OnHit(key)
	} else {
		s.policy.OnSet(key)
	}
	return true, evicted
}

func (s *Shard[K, V]) Get(key K) (V, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.store[key]
	if !ok {
		var zero V
		return zero, false
	}

	now := s.now()
	if s.defaultTTL != 0 && (entry.expiresAt.Before(now) || entry.expiresAt.Equal(now)) {
		delete(s.store, key)
		s.policy.OnDel(key)
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
			if _, ok := s.store[victim]; ok {
				delete(s.store, victim)
				s.policy.OnDel(victim)
			}
		}
	}

	s.policy.OnHit(key)
	return entry.val, true
}

// no policy effects
func (s *Shard[K, V]) Peek(key K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.store[key]
	if !ok {
		var zero V
		return zero, false
	}
	return entry.val, true
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

func (s *Shard[K, V]) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()

	clear(s.store)
	s.policy.Reset()
	s.expiry.Reset()
}

func (s *Shard[K, V]) Equals(o *Shard[K, V]) bool {
	sPtype, sKtype := s.policy.Type()
	oPtype, oKtype := o.policy.Type()
	return reflect.DeepEqual(s.store, o.store) &&
		s.cap == o.cap &&
		sPtype == oPtype &&
		sKtype == oKtype
}

func (s *Shard[K, V]) setNow(now func() time.Time) {
	s.now = now
	s.expiry.SetNow(now)
}

func (s *Shard[K, V]) validate() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.expiry.Len() < len(s.store) {
		return errors.New("expiry out of sync")
	}
	if len(s.store) != s.policy.Len() {
		return errors.New("policy out of sync")
	}
	return nil
}
