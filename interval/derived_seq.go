// Find the derived sequence of graphs, G^1...G^n, based on the intervals of a
// control flow graph G, as described in figure 2 "Derived Sequence Algorithm"
// in C. Cifuentes, "A Structuring Algorithm for Decompilation", 1993.

package interval

import (
	"fmt"

	"github.com/graphism/exp/cfg"
	"gonum.org/v1/gonum/graph"
)

// TODO: Is there a difference between immedPreds and immedPred?

// DerivedSeq returns the derived sequence of graphs, G^1...G^n, based on the
// intervals of the given control flow graph G, and the associated unique sets
// of intervals, 𝓘^1...𝓘^n.
func DerivedSeq(g *cfg.Graph) ([]*cfg.Graph, [][]*Interval) {
	// G^1 = G
	g.SetDOTID("G1")
	Gs := []*cfg.Graph{g}
	// 𝓘^1 = intervals(G^1)
	IIs := [][]*Interval{Intervals(g)}
	// i = 2
	i := 1 // 0-indexed.
	// repeat /* construction of G^i */
	for {
		//    make each interval of G^{i-1} a node in G^i
		GNew := Gs[i-1]
		for j, I := range IIs[i-1] {
			delNodes := make(map[string]bool)
			for _, n := range I.Nodes() {
				delNodes[dotID(n)] = true
			}
			newName := fmt.Sprintf("G%d_I%d", i, j+1)

			// The collapsed node n of an interval I(h) has the immediate
			// predecessors of h not part of the interval.
			//
			//    immedPreds(n) n ∈ G^i = immedPreds(h) : immedPred(h) ∉ I^{i-1}(h)
			//
			// The collapsed node n of an interval I(h) has the immediate
			// successors of the exit nodes of I(h) not part of the interval.
			//
			//    (a, b) ∈ G^i iff ∃ n ∈ I^{i-1}(h) and m = header(I^{i-1}(m)) : (m, n) ∈ G^{i-1}

			// TODO: Figure out why the paper uses
			//    ∃ n ∈ I^{i-1}(h)
			// rather than
			//    ∃ n ∉ I^{i-1}(h)
			GNew = cfg.Merge(GNew, delNodes, newName)
			newNode, ok := GNew.NodeWithName(newName)
			if !ok {
				panic(fmt.Errorf("unable to locate collapsed node %q", newName))
			}
			// TODO: Validate that this is a valid way to track the switch header
			// node, when collapsing the nodes of an interval.
			newNode.SwitchHead = I.h.SwitchHead
		}
		GNew.SetDOTID(fmt.Sprintf("G%d", i+1))
		if len(GNew.Nodes()) == len(Gs[i-1].Nodes()) {
			break
		}
		Gs = append(Gs, GNew)
		//    𝓘^i = intervals(G^i)
		IIs = append(IIs, Intervals(Gs[i]))
		// until
		//    G^i == G^{i-1}
		i++
	}
	return Gs, IIs
}

// ### [ Helper functions ] ####################################################

// dotID returns the DOT ID of the given node.
func dotID(n graph.Node) string {
	return n.(*cfg.Node).DOTID()
}
