package core

import (
	"sync/atomic"
)

type Stats struct {
	Hits      atomic.Uint64
	Misses    atomic.Uint64
	Evictions atomic.Uint64
	Deletes   atomic.Uint64
	Flushes   atomic.Uint64
}

type StatsSnapshot struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
	Deletes   uint64
	Flushes   uint64
}
