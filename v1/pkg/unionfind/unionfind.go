// Package unionfind provides disjoint-set data structure for dynamic clustering.
//
// Union and Find both amortized O(α(n)) ≈ O(1) with path compression + union by rank.
// Used to merge keyword clusters incrementally without reprocessing.
//
// Usage:
//
//	uf := unionfind.New(1000)
//	uf.Union(0, 1)  // merge clusters
//	root := uf.Find(0)
package unionfind

// UF is a Union-Find (disjoint-set) with path compression and union by rank.
type UF struct {
	Parent []int
	Rank   []int
	Size   []int
}

// New creates a UF with n elements (0..n-1).
func New(n int) *UF {
	uf := &UF{
		Parent: make([]int, n),
		Rank:   make([]int, n),
		Size:   make([]int, n),
	}
	for i := range uf.Parent {
		uf.Parent[i] = i
		uf.Size[i] = 1
	}
	return uf
}

// Find returns the root of x with path compression.
func (uf *UF) Find(x int) int {
	if uf.Parent[x] != x {
		uf.Parent[x] = uf.Find(uf.Parent[x])
	}
	return uf.Parent[x]
}

// Union merges the sets containing x and y. Returns true if merged.
func (uf *UF) Union(x, y int) bool {
	rx, ry := uf.Find(x), uf.Find(y)
	if rx == ry {
		return false
	}
	if uf.Rank[rx] < uf.Rank[ry] {
		rx, ry = ry, rx
	}
	uf.Parent[ry] = rx
	uf.Size[rx] += uf.Size[ry]
	if uf.Rank[rx] == uf.Rank[ry] {
		uf.Rank[rx]++
	}
	return true
}

// Connected reports whether x and y are in the same set.
func (uf *UF) Connected(x, y int) bool {
	return uf.Find(x) == uf.Find(y)
}

// ClusterSize returns the size of the cluster containing x.
func (uf *UF) ClusterSize(x int) int {
	return uf.Size[uf.Find(x)]
}
