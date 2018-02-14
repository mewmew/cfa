//+build ignore

// TODO: Remove this file once the cfa package has matured.

package interval

import (
	"github.com/graphism/exp/cfg"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
)

// === [ "A Structuring Algorithm for Decompilation", 1993 ] ===================

// --- [ Figure 7 "Loop structuring algorithm" ] -------------------------------

// Structure the loops of a control flow graph, as described in figure 7 "Loop
// structuring algorithm" in C. Cifuentes, "A Structuring Algorithm for
// Decompilation", 1993.

// structLoop2 structures loops in the given control flow graph.
func structLoop2(g *cfg.Graph) {
	Gs, IIs := DerivedSeq(g)
	// for (G^i = G^1...G^n)
	for i := range Gs {
		//    for (I^i(h_j) = I^1(h_1)...I^m(h_m))
		dom := path.Dominators(Gs[i].Entry(), Gs[i])
		for _, I := range IIs[i] {
			//       if (h_j has a back edge, (x, h_j))
			for _, x := range Gs[i].To(I.h) {
				if I.Has(x) {
					//          if (x.inLoop == false)
					xx := node(x)
					if xx.LoopHead == nil {
						xx.LoopHead = I.h // TODO: Remove?
						//             find all nodes in loop
						nodes := nodesInLoop(I, xx, dom)
						for _, n := range nodes {
							//             flag inLoop for all these nodes
							n.LoopHead = I.h
						}
					} else {
						//          else
						//             flag h_j.label
						// TODO: flag use of goto label?
					}
					//          fi
				}
				//       fi
			}
		}
		//    endFor
	}
	// endFor
}

// --- [ Figure 6 "Finding nodes in a loop" ] ----------------------------------

// Locate nodes of the loop in the interval I(h), as described in figure 6
// "Finding nodes in a loop" in C. Cifuentes, "A Structuring Algorithm for
// Decompilation", 1993.

// nodesInLoop returns the nodes of the loop in the interval I(h).
func nodesInLoop(I *Interval, x *cfg.Node, dom path.DominatorTree) []*cfg.Node {
	// nodesInLoop = {revPost(h)}
	nodesInLoop := map[graph.Node]bool{
		I.h: true,
	}
	// for (i = revPost(h)+1 ... revPost(a)-1)
	for _, i := range cfg.SortByPost(I.Nodes()) {
		//    if ((revPost(immedDom(i)) ∈ nodesInLoop) and (i ∈ nodesInInterval))
		if nodesInLoop[dom.DominatorOf(i)] {
			//       nodesInLoop = nodesInLoop + {i}
			nodesInLoop[i] = true
		}
		//    fi
	}
	// endFor
	// nodesInLoop = nodesInLoop + {revPost(a)}
	nodesInLoop[x] = true
	var nodes []*cfg.Node
	for n := range nodesInLoop {
		nodes = append(nodes, node(n))
	}
	return nodes
}
