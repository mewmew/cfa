// Package interval implements control flow recovery based on the interval
// method, as described in C. Cifuentes, "A Structuring Algorithm for
// Decompilation", 1993.
package interval

import (
	"github.com/graphism/exp/cfg"
	"gonum.org/v1/gonum/graph"
)

// Analyze analyzes the given control flow graph using the interval method.
func Analyze(g *cfg.Graph) {
}

// Intervals returns the intervals of the given control flow graph.
func Intervals(g *cfg.Graph) []*Interval {
	// 𝓘 = {}
	var Is []*Interval
	// H = {h}
	H := map[graph.Node]bool{
		g.Entry(): true,
	}
	// for (all unprocessed n ∈ H) do
	for n := range H {
		// I(n) = {n}
		I := newInterval(g, n)
		// repeat
		//    I(n) = I(n) + {m ∈ G : ∀ p = immedPred(m), p ∈ I(n) }
		// until
		//    no more nodes can be added to I(n)
		for {
			added := false
			for _, m := range g.Nodes() {
				if I.nodes[m] {
					continue
				}
				if I.containsAllPreds(m) {
					I.nodes[m] = true
					added = true
				}
			}
			if !added {
				break
			}
		}
		// H = H + {m ∈ G : m ∉ H and m ∉ I(n) and (∃ p = immedPred(m) : p ∈ I(n))}
		for _, m := range g.Nodes() {
			if H[m] {
				continue
			}
			if I.Has(m) {
				continue
			}
			for _, p := range g.To(m) {
				if I.Has(p) {
					H[m] = true
					break
				}
			}
		}
		// 𝓘 = 𝓘 + I(n)
		Is = append(Is, I)
	}
	return Is
}

// TODO: consider changing from `nodes map[graph.Node]bool` to `nodes
// []graph.Node` sorted in reverse post-order in Interval.

// Interval is the interval I(h) with header node h, the maximal, single-entry
// subgraph in which h is the only entry node and in which all closed paths
// contain h.
type Interval struct {
	// Control flow graph containing I(h).
	g *cfg.Graph
	// Header node.
	h graph.Node
	// Nodes of the interval.
	nodes map[graph.Node]bool
}

// Has reports whether the node exists within the interval.
func (I *Interval) Has(n graph.Node) bool {
	return I.nodes[n]
}

// Nodes returns the nodes of the interval.
func (I *Interval) Nodes() []graph.Node {
	var nodes []graph.Node
	for node := range I.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// From returns all nodes that can be reached directly from the given node.
func (I *Interval) From(n graph.Node) []graph.Node {
	return I.g.From(n)
}

// To returns all nodes that can reach directly to the given node.
func (I *Interval) To(n graph.Node) []graph.Node {
	return I.g.To(n)
}

// newInterval returns a new interval I(h) with header node h.
func newInterval(g *cfg.Graph, h graph.Node) *Interval {
	return &Interval{
		g: g,
		h: h,
		nodes: map[graph.Node]bool{
			h: true,
		},
	}
}

// containsAllPreds reports whether the interval I(h) contains all the immediate
// predecessors of n.
func (I *Interval) containsAllPreds(n graph.Node) bool {
	for _, p := range I.From(n) {
		if !I.Has(p) {
			return false
		}
	}
	return true
}
