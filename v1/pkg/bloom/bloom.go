// Package bloom provides a space-efficient probabilistic set for keyword dedup.
//
// O(1) insertion and lookup with configurable false-positive rate (~1% default).
// For 10M keywords: ~10 MB memory (vs ~400 MB for map[string]bool).
//
// Usage:
//
//	f := bloom.New(10_000_000, 0.01)
//	f.Add("automação de atendimento")
//	if f.Contains("automação de atendimento") { ... }
package bloom

import "hash/fnv"

// Filter is a Bloom Filter for keyword deduplication.
type Filter struct {
	bits []uint64
	k    int    // number of hash functions
	m    uint64 // size in bits
}

// New creates a Bloom Filter for n expected items with p false-positive rate.
func New(n int, p float64) *Filter {
	m := uint64(float64(-n) * 1.44 * p) // approximate optimal m
	k := int(0.7 * float64(m) / float64(n))
	if k < 1 {
		k = 1
	}
	words := (m + 63) / 64
	return &Filter{
		bits: make([]uint64, words),
		k:    k,
		m:    m,
	}
}

// Add inserts a key into the filter.
func (f *Filter) Add(key string) {
	h1, h2 := f.hash(key)
	for i := 0; i < f.k; i++ {
		idx := (h1 + uint64(i)*h2) % f.m
		f.bits[idx/64] |= 1 << (idx % 64)
	}
}

// Contains reports whether key is probably in the filter.
func (f *Filter) Contains(key string) bool {
	h1, h2 := f.hash(key)
	for i := 0; i < f.k; i++ {
		idx := (h1 + uint64(i)*h2) % f.m
		if f.bits[idx/64]&(1<<(idx%64)) == 0 {
			return false
		}
	}
	return true
}

func (f *Filter) hash(key string) (uint64, uint64) {
	h := fnv.New64a()
	h.Write([]byte(key))
	h1 := h.Sum64()
	h.Reset()
	h.Write([]byte(key + "salt"))
	return h1, h.Sum64()
}
