package internal

import (
	"context"
	"time"
)

type Janitor[K comparable, V any] struct {
	shard  *Shard[K, V]
	cancel context.CancelFunc
	result chan uint64
}

func StartJanitor[K comparable, V any](s *Shard[K, V], initialDelay time.Duration) *Janitor[K, V] {
	janitor := &Janitor[K, V]{
		shard:  s,
		result: make(chan uint64),
	}

	ctx, cancel := context.WithCancel(context.Background())
	janitor.cancel = cancel

	go janitor.run(ctx, initialDelay)

	return janitor
}

func (j *Janitor[K, V]) Start(initialDelay time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	j.cancel = cancel

	go j.run(ctx, initialDelay)
}

func (j *Janitor[K, V]) run(ctx context.Context, initialDelay time.Duration) {
	timer := time.NewTimer(initialDelay)
	defer timer.Stop()

	var expired uint64

	for {
		select {
		case <-ctx.Done():
			// final sweep before exiting for consistency
			j.shard.mu.Lock()
			for j.shard.expiry.HasExpired() {
				victim := j.shard.expiry.PopMin().K
				if _, ok := j.shard.Store[victim]; ok {
					delete(j.shard.Store, victim)
					j.shard.Policy.OnDel(victim)
					expired++
				}
			}
			j.shard.mu.Unlock()

			j.result <- expired
			close(j.result)
			return
		case <-timer.C:
			j.shard.mu.Lock()
			// cleanup
			for j.shard.expiry.HasExpired() {
				victim := j.shard.expiry.PopMin().K
				if _, ok := j.shard.Store[victim]; ok {
					delete(j.shard.Store, victim)
					j.shard.Policy.OnDel(victim)
					expired++
				}
			}

			// determine next sweep
			var d time.Duration
			if e, ok := j.shard.expiry.Peek(); ok {
				d = max(time.Until(e.ExpiresAt), 0)
			} else {
				d = initialDelay
			}
			j.shard.mu.Unlock()

			// drain timer
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(d)
		}
	}
}

func (j *Janitor[K, V]) Stop() (expired uint64) {
	j.cancel()
	return <-j.result
}
