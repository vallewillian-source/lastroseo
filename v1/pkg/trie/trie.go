// Package trie implements a prefix tree for Google Autocomplete suggestions.
//
// Insert and prefix search both O(L) where L = keyword length.
// For ~1M suggestions: ~200 MB (runes + maps).
//
// Usage:
//
//	t := trie.New()
//	t.Insert("automação de atendimento", 100)
//	results := t.PrefixSearch("automação", 10)
package trie

import "sort"

// Node is a trie node.
type Node struct {
	Children map[rune]*Node
	Weight   float64
	IsEnd    bool
}

// Trie is a prefix tree.
type Trie struct {
	Root *Node
}

// New returns an empty Trie.
func New() *Trie {
	return &Trie{Root: &Node{Children: make(map[rune]*Node)}}
}

// Insert adds a word with a weight.
func (t *Trie) Insert(word string, weight float64) {
	node := t.Root
	for _, r := range word {
		if _, ok := node.Children[r]; !ok {
			node.Children[r] = &Node{Children: make(map[rune]*Node)}
		}
		node = node.Children[r]
	}
	node.IsEnd = true
	node.Weight = weight
}

// PrefixSearch finds up to `limit` words matching the prefix, sorted by weight desc.
func (t *Trie) PrefixSearch(prefix string, limit int) []Result {
	node := t.Root
	for _, r := range prefix {
		if child, ok := node.Children[r]; ok {
			node = child
		} else {
			return nil
		}
	}
	var results []Result
	t.collect(node, prefix, &results)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Weight > results[j].Weight
	})
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results
}

// Result is a prefix search result.
type Result struct {
	Word   string
	Weight float64
}

func (t *Trie) collect(node *Node, prefix string, results *[]Result) {
	if node.IsEnd {
		*results = append(*results, Result{Word: prefix, Weight: node.Weight})
	}
	for r, child := range node.Children {
		t.collect(child, prefix+string(r), results)
	}
}
