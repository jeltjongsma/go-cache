package hasher

import (
	"encoding"
	"encoding/binary"
	"fmt"
	"hash/maphash"
	"math"
)

type Hasher[K comparable] struct {
	seed      maphash.Seed
	keyWriter KeyWriter[K]
}

func NewHasher[K comparable](keyWriter KeyWriter[K]) *Hasher[K] {
	if keyWriter == nil {
		keyWriter = &DefaultKeyWriter[K]{}
	}
	return &Hasher[K]{
		seed:      maphash.MakeSeed(),
		keyWriter: keyWriter,
	}
}

func (h *Hasher[K]) Hash(k K) uint64 {
	var hash maphash.Hash
	hash.SetSeed(h.seed)
	h.keyWriter.WriteKey(&hash, k)
	return hash.Sum64()
}

type KeyWriter[K any] interface {
	WriteKey(h *maphash.Hash, k K)
}

type DefaultKeyWriter[K any] struct{}

func (DefaultKeyWriter[K]) WriteKey(h *maphash.Hash, k K) {
	switch v := any(k).(type) {
	case string:
		h.WriteString(v) // no extra alloc
	case bool:
		if v {
			h.WriteByte(1)
		} else {
			h.WriteByte(0)
		}
	case int:
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], uint64(v))
		h.Write(buf[:])
	case int32:
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], uint32(v))
		h.Write(buf[:])
	case int64:
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], uint64(v))
		h.Write(buf[:])
	case uint, uint32:
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], uint32(any(v).(uint32)))
		h.Write(buf[:])
	case uint64:
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], v)
		h.Write(buf[:])
	case float64:
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(v))
		h.Write(buf[:])
	case float32:
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(v))
		h.Write(buf[:])
	case [16]byte: // common fixed-size IDs
		h.Write(v[:])
	case encoding.BinaryMarshaler: // stdlib interface
		if b, err := v.MarshalBinary(); err == nil {
			h.Write(b)
		} else {
			// fall back: write error code path marker to keep determinism
			h.WriteByte(0xff)
		}
	case fmt.Stringer: // last-resort (allocates)
		h.WriteString(v.String())
	default:
		// Force the user to provide a custom KeyWriter for structs/complex types.
		panic("no default hashing for this key type; supply a KeyWriter[K]")
	}
}
