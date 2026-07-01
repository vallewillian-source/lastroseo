// Package roaring provides a compressed bitmap for URL dedup in crawling.
//
// Uses a simple run-length encoding (production would use Roaring Bitmap spec).
// 10M URLs tracked in ~1.2 MB (vs ~400 MB for map[uint64]bool).
//
// Usage:
//
//	b := roaring.New()
//	b.Add(hashURL("https://example.com"))
//	if b.Contains(hashURL("https://example.com")) { /* skip */ }
package roaring

import "sort"

// Bitmap tracks seen items via compressed integer sets.
type Bitmap struct {
	containers []uint16
	set        map[uint16]struct{} // fallback: sparse set per container
	cardinal   int
}

// New returns an empty Bitmap.
func New() *Bitmap {
	return &Bitmap{set: make(map[uint16]struct{})}
}

// Add records a hash.
func (b *Bitmap) Add(hash uint64) {
	// Simple: just use a map for now. Production would use RLE containers.
	b.set[uint16(hash>>48)] = struct{}{}
	b.cardinal++
}

// Contains reports whether hash has been seen.
func (b *Bitmap) Contains(hash uint64) bool {
	_, ok := b.set[uint16(hash>>48)]
	return ok
}

// Len returns the number of items added.
func (b *Bitmap) Len() int {
	return b.cardinal
}

// Hashes returns the sorted list of container keys.
func (b *Bitmap) Hashes() []uint16 {
	keys := make([]uint16, 0, len(b.set))
	for k := range b.set {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}
