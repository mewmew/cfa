package interval

import (
	"fmt"
	"log"
	"os"

	"github.com/graphism/exp/cfg"
	"github.com/mewkiz/pkg/term"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
)

var (
	// dbg represents a logger with the "interval:" prefix, which logs debug
	// messages to standard error.
	dbg = log.New(os.Stderr, term.BlueBold("interval:")+" ", 0)
	// warn represents a logger with the "interval:" prefix, which logs warnings
	// to standard error.
	warn = log.New(os.Stderr, term.RedBold("interval:")+" ", 0)
)

// Analyze analyzes the given control flow graph using the interval method.
func Analyze(g *cfg.Graph) *Primitives {
	prims := NewPrimitives()
	// Structure loops.
	structLoop(g, prims)
	// TODO: Structure if-statements.
	return prims
}

// === [ DCC ] =================================================================

// --- [ structLoops ] ---------------------------------------------------------

// structLoop structures loops in the given control flow graph.
func structLoop(g *cfg.Graph, prims *Primitives) {
	// Note, the call to DerivedSeq initiates the reverse post-order number of
	// each node.
	// For all derived sequences G_i.
	Gs, IIs := DerivedSeq(g)
	for i, Gi := range Gs {
		// TODO: Remove when cfa has matured. Useful for debugging.
		//if err := ioutil.WriteFile(fmt.Sprintf("G_%d.dot", i), []byte(Gs[i].String()), 0644); err != nil {
		//	log.Fatalf("%+v", err)
		//}
		// For all intervals I_i of G_i.
		dom := path.Dominators(Gi.Entry(), Gi)
		for j, I := range IIs[i] {
			// Record interval information.
			intervalName := fmt.Sprintf("G%d_I%d", i, j+1)
			intervalNodes := nodeNames(cfg.SortByRevPost(I.Nodes()))
			if prev, ok := prims.Intervals[intervalName]; ok {
				panic(fmt.Errorf("interval with name %q already present; prev nodes %v, new nodes %v", intervalName, prev, intervalNodes))
			}
			prims.Intervals[intervalName] = intervalNodes
			// Find greatest enclosing back edge (if any).
			var latch *cfg.Node
			for _, pred := range cfg.SortByRevPost(Gi.To(I.h)) {
				// TODO: Remove when cfa has matured. Useful for debugging.
				//dbg.Printf("pred of %v: %v\n", I.h.DOTID(), pred.DOTID())
				//dbg.Printf("I.Has(%v)=%v\n", pred.DOTID(), I.Has(pred))
				//dbg.Printf("isBackEdge(%v, %v)=%v\n", pred.DOTID(), I.h.DOTID(), isBackEdge(pred, I.h))
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
				dbg.Println("located latch node:", latch.DOTID())
				// TODO: Handle case statements when implemented.
				// Check that the node doesn't belong to another loop.
				if latch.LoopHead == nil {
					I.h.Latch = latch
					loop := findNodesInLoop(Gi, I, latch, dom)
					// Record loop information.
					prims.Loops = append(prims.Loops, loop)
					latch.IsLatch = true // TODO: Remove if not needed.
				}
			}
		}
	}
}

// --- [ findNodesInLoop ] -----------------------------------------------------

// findNodesInLoop locates the nodes in the loop (latch, I.h) and determines the
// type of the loop.
func findNodesInLoop(g *cfg.Graph, I *Interval, latch *cfg.Node, dom path.DominatorTree) *Loop {
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

	// Collect information about located loop.
	follow := ""
	if I.h.LoopFollow != nil {
		follow = I.h.LoopFollow.DOTID()
	}
	loop := &Loop{
		Type:   I.h.LoopType,
		Head:   I.h.DOTID(),
		Latch:  latch.DOTID(),
		Follow: follow,
	}
	var ns []graph.Node
	for n := range loopNodes {
		ns = append(ns, n)
	}
	for _, n := range cfg.SortByRevPost(ns) {
		loop.Nodes = append(loop.Nodes, n.DOTID())
	}
	return loop
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

// ### [ Helper functions ] ####################################################

// nodeNames returns the DOTID node names of the given nodes.
func nodeNames(nodes []*cfg.Node) []string {
	var ids []string
	for _, n := range nodes {
		ids = append(ids, n.DOTID())
	}
	return ids
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
