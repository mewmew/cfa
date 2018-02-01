// Find the unique set of intervals of a control flow graph, as described in
// figure 1 "Interval Algorithm" in C. Cifuentes, "A Structuring Algorithm for
// Decompilation", 1993.

package interval

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/graphism/exp/cfg"
	"gonum.org/v1/gonum/graph"
)

// Intervals returns the unique set of intervals of the given control flow
// graph.
func Intervals(g *cfg.Graph) []*Interval {
	// 𝓘 = {}
	var Is []*Interval
	// H = {h}
	H := newQueue()
	H.push(g.Entry())
	// for (all unprocessed n ∈ H) do
	for !H.empty() {
		n := H.pop()
		// I(n) = {n}
		I := newInterval(g, n)
		// repeat
		//    I(n) = I(n) + {m ∈ G : ∀ p = immedPred(m), p ∈ I(n) }
		// until
		//    no more nodes can be added to I(n)
		for added := true; added; {
			added = false
			for _, m := range g.Nodes() {
				if I.nodes[m] {
					continue
				}
				// Check that m has at least one predecessor and that all
				// predecessors are in I(n).
				if I.containsAllPreds(m) {
					I.nodes[m] = true
					added = true
				}
			}
		}
		// H = H + {m ∈ G : m ∉ H and m ∉ I(n) and (∃ p = immedPred(m) : p ∈ I(n))}
		for _, m := range g.Nodes() {
			if H.has(m) {
				continue
			}
			if I.Has(m) {
				continue
			}
			for _, p := range g.To(m) {
				if I.Has(p) {
					H.push(m)
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

// String returns a string representation of the interval.
func (I *Interval) String() string {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "I(%v): {", I.h.(*cfg.Node).DOTID())
	var ids []string
	for n := range I.nodes {
		id := n.(*cfg.Node).DOTID()
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for i, id := range ids {
		if i != 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(id)
	}
	buf.WriteString("}")
	return buf.String()
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
// predecessors of n and n has at least one predecessor.
func (I *Interval) containsAllPreds(n graph.Node) bool {
	preds := I.To(n)
	if len(preds) == 0 {
		// Ignore nodes without predecessors (e.g. entry node); otherwise they
		// would be added to every interval.
		return false
	}
	for _, p := range preds {
		if !I.Has(p) {
			return false
		}
	}
	return true
}

// ### [ Helper functions ] ####################################################

// A queue is a FIFO queue of nodes which keeps track of all nodes that has been
// in the queue.
type queue struct {
	// List of nodes in queue.
	l []graph.Node
	// Current position in queue.
	i int
}

// newQueue returns a new FIFO queue.
func newQueue() *queue {
	return &queue{
		l: make([]graph.Node, 0),
	}
}

// push appends the given node to the end of the queue if it has not been in the
// queue before.
func (q *queue) push(n graph.Node) {
	if !q.has(n) {
		q.l = append(q.l, n)
	}
}

// has reports whether the given node is present in the queue or has been
// present before.
func (q *queue) has(n graph.Node) bool {
	for _, m := range q.l {
		if n == m {
			return true
		}
	}
	return false
}

// pop pops and returns the first node of the queue.
func (q *queue) pop() graph.Node {
	if q.empty() {
		panic("invalid call to pop; empty queue")
	}
	n := q.l[q.i]
	q.i++
	return n
}

// empty reports whether the queue is empty.
func (q *queue) empty() bool {
	return len(q.l[q.i:]) == 0
}
