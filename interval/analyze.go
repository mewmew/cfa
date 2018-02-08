package interval

import (
	"fmt"

	"github.com/graphism/exp/cfg"
	"github.com/kr/pretty"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
)

// Analyze analyzes the given control flow graph using the interval method.
func Analyze(g *cfg.Graph) {
	dom := path.Dominators(g.Entry(), g)
	structLoop(g, dom)
}

// === [ DCC ] =================================================================

// --- [ structLoops ] ---------------------------------------------------------

// structLoop structures loops in the given control flow graph.
func structLoop(g *cfg.Graph, dom path.DominatorTree) {
	// For all derived sequences G_i.
	Gs, IIs := DerivedSeq(g)
	for i := range Gs {
		// For all intervals I_i of G_i.
		for _, I := range IIs[i] {
			// Find greatest enclosing back edge (if any).
			var latch *cfg.Node
			for _, pred := range cfg.SortByRevPost(g.To(I.h)) {
				if I.Has(pred) && isBackEdge(pred, I.h) {
					if latch == nil {
						latch = pred
					} else if pred.RevPost > latch.RevPost {
						latch = pred
					}
				}
			}
			// Find nodes in the loop and the type of the loop.
			if latch != nil {
				// TODO: Handle case statements when implemented.
				// Check that the node doesn't belong to another loop.
				if latch.LoopHead == nil {
					I.h.Latch = latch
					findNodesInLoop(g, I, latch, dom)
					latch.IsLatch = true
				}
			}
		}
	}
	// TODO: Remove debug output. Instead collect information from control flow
	// analysis, for each derived graph G^i.
	for i := range Gs {
		fmt.Printf("--- [ Gs[%d] ] --------------------------\n", i)
		fmt.Println()
		for _, n := range cfg.SortByRevPost(Gs[i].Nodes()) {
			pretty.Println("node:", n)
			if n.LoopHead != nil {
				pretty.Println("   LoopHead:", n.LoopHead.DOTID())
			}
			if n.Latch != nil {
				pretty.Println("   Latch:", n.Latch.DOTID())
			}
			if n.LoopFollow != nil {
				pretty.Println("   LoopFollow:", n.LoopFollow.DOTID())
			}
		}
		fmt.Println()
	}
}

// --- [ findNodesInLoop ] -----------------------------------------------------

// findNodesInLoop locates the nodes in the loop (latch, I.h) and determines the
// type of the loop.
func findNodesInLoop(g *cfg.Graph, I *Interval, latch *cfg.Node, dom path.DominatorTree) {
	// Flag nodes in loop headed by head (except header node).
	I.h.LoopHead = I.h
	loopNodes := make(map[*cfg.Node]bool)
	for _, n := range cfg.SortByRevPost(I.Nodes()) {
		if n == I.h {
			// skip header node.
			continue
		}
		if n == latch {
			// skip latch node.
			break
		}
		if immedDom := dom.DominatorOf(n); loopNodes[node(immedDom)] {
			loopNodes[n] = true
			if n.LoopHead == nil {
				n.LoopHead = I.h
			}
		}
	}
	latch.LoopHead = I.h
	if latch != I.h {
		loopNodes[latch] = true
	}

	// Determine the type of the loop and the follow node.
	headSuccs := g.From(I.h)
	latchSuccs := g.From(latch)
	switch len(latchSuccs) {
	// latch = 2-way
	case 2:
		latchTrueTarget, latchFalseTarget := g.TrueTarget(latch), g.FalseTarget(latch)
		switch {
		// latch is 2-way and all successors of head is within the loop.
		case len(headSuccs) == 2 || latch == I.h:
			if latch == I.h || loopNodes[node(headSuccs[0])] && loopNodes[node(headSuccs[1])] {
				I.h.LoopType = cfg.LoopTypePostTest
				if latchTrueTarget == I.h {
					I.h.LoopFollow = latchFalseTarget
				} else {
					I.h.LoopFollow = latchTrueTarget
				}
				// head has successor outside of the loop.
			} else {
				headTrueTarget, headFalseTarget := g.TrueTarget(I.h), g.FalseTarget(I.h)
				I.h.LoopType = cfg.LoopTypePreTest
				if loopNodes[headTrueTarget] {
					I.h.LoopFollow = headFalseTarget
				} else {
					I.h.LoopFollow = headTrueTarget
				}
			}
		// head = anything besides 2-way, latch = 2-way
		default:
			I.h.LoopType = cfg.LoopTypePostTest
			if latchTrueTarget == I.h {
				I.h.LoopFollow = latchFalseTarget
			} else {
				I.h.LoopFollow = latchTrueTarget
			}
		}
	// latch = 1-way
	case 1:
		if len(headSuccs) == 2 {
			I.h.LoopType = cfg.LoopTypePreTest
			n := latch
			headTrueTarget, headFalseTarget := g.TrueTarget(I.h), g.FalseTarget(I.h)
			trueTarget := headTrueTarget
			falseTarget := headFalseTarget
			for {
				if n == trueTarget {
					I.h.LoopFollow = falseTarget
					break
				} else if n == falseTarget {
					I.h.LoopFollow = trueTarget
					break
				}
				// Check if the follow node couldn't be found, the it is a strangely
				// formed loop, so it is safer to consider it an endless loop.
				if n.RevPost <= I.h.RevPost {
					I.h.LoopType = cfg.LoopTypeEndless
					// Note, the follow node is yet to be located.
					break
				}
				n = node(dom.DominatorOf(n))
			}
			if n.RevPost > I.h.RevPost {
				I.h.LoopFollow.LoopHead = nil
			}
		} else {
			I.h.LoopType = cfg.LoopTypeEndless
			// Note, the follow node is yet to be located.
		}
	default:
		panic(fmt.Errorf("support for latch node with %d successors not yet implemented", len(latchSuccs)))
	}
}

// --- [ isBackEdge ] ----------------------------------------------------------

// isBackEdge reports whether (p, s) is a back edge; if the successor s was
// visited before the predecessor p during a depth first traversal of the graph.
func isBackEdge(p, s *cfg.Node) bool {
	if p.Pre >= s.Pre {
		// TODO: Check if needed; and if it is better placed somewhere else as the
		// isBackEdge function name does not communicate that it also alters the
		// node members.
		s.NBackEdges++
		return true
	}
	return false
}

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
