// Package kdtree provides nearest-neighbor search for 384-dim embedding vectors.
//
// O(log n) nearest-neighbor queries. Used to find semantically similar keywords
// for keyword expansion and clustering.
//
// Usage:
//
//	tree := kdtree.New(384)
//	tree.Insert(embedding1, 0)
//	tree.Insert(embedding2, 1)
//	neighbors := tree.Nearest(embedding1, 5)
package kdtree

import "math"

// KDTree is a k-dimensional tree for embedding vectors.
type KDTree struct {
	Root *Node
	Dims int
}

// Node is a KDTree node.
type Node struct {
	Point    []float64
	ID       int
	Left     *Node
	Right    *Node
	SplitDim int
}

// New creates a KDTree for vectors of `dims` dimensions.
func New(dims int) *KDTree {
	return &KDTree{Dims: dims}
}

// Insert adds a point with the given ID.
func (t *KDTree) Insert(point []float64, id int) {
	t.Root = t.insert(t.Root, point, id, 0)
}

func (t *KDTree) insert(node *Node, point []float64, id, depth int) *Node {
	if node == nil {
		return &Node{Point: point, ID: id, SplitDim: depth % t.Dims}
	}
	dim := node.SplitDim
	if point[dim] < node.Point[dim] {
		node.Left = t.insert(node.Left, point, id, depth+1)
	} else {
		node.Right = t.insert(node.Right, point, id, depth+1)
	}
	return node
}

// Neighbor holds a nearest-neighbor result.
type Neighbor struct {
	ID       int
	Distance float64
}

// Nearest returns the k nearest neighbors of the target point.
func (t *KDTree) Nearest(target []float64, k int) []Neighbor {
	var results []Neighbor
	t.nearest(t.Root, target, k, &results)
	return results
}

func (t *KDTree) nearest(node *Node, target []float64, k int, results *[]Neighbor) {
	if node == nil {
		return
	}
	dist := euclidean(node.Point, target)
	*results = append(*results, Neighbor{ID: node.ID, Distance: dist})

	dim := node.SplitDim
	var near, far *Node
	if target[dim] < node.Point[dim] {
		near, far = node.Left, node.Right
	} else {
		near, far = node.Right, node.Left
	}
	t.nearest(near, target, k, results)
	if len(*results) < k || math.Abs(target[dim]-node.Point[dim]) < (*results)[len(*results)-1].Distance {
		t.nearest(far, target, k, results)
	}
}

func euclidean(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		d := a[i] - b[i]
		sum += d * d
	}
	return math.Sqrt(sum)
}
