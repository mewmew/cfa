package interval

import (
	"fmt"
	"log"
	"os"

	"github.com/graphism/exp/cfg"
	"github.com/kr/pretty"
	"github.com/mewkiz/pkg/term"
	"github.com/mewmew/cfa/primitive"
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
func Analyze(g *cfg.Graph) *primitive.Primitives {
	prims := primitive.NewPrimitives()
	dom := path.Dominators(g.Entry(), g)
	// Structure switch statements.
	structSwitch(g, prims, dom)
	// Structure loops.
	structLoop(g, prims)
	// Structure if-statements.
	structIf(g, prims, dom)
	return prims
}

// === [ DCC ] =================================================================

// --- [ structCase ] ----------------------------------------------------------

// structSwitch structures switch statements in the given control flow graph.
func structSwitch(g *cfg.Graph, prims *primitive.Primitives, dom path.DominatorTree) {
	// Search for case nodes in reverse post-order.
	for _, n := range cfg.SortByRevPost(g.Nodes()) {
		headSuccs := cfg.SortByRevPost(g.From(n))
		if len(headSuccs) > 2 {
			// Switch header node.
			head := n
			// Switch follow node.
			var follow *cfg.Node
			switchNodes := make(map[*cfg.Node]bool)
			// Find descendant node which has the current header node as immediate
			// predecessor, and is not a successor.
			//
			//     head
			//    / | \
			//    \ | /
			//    follow
			immedDoms := cfg.SortByRevPost(dom.DominatedBy(head))
			for _, immedDom := range immedDoms {
				if isSuccessor(g, immedDom, head) {
					// Skip immediate successors of the header node.
					continue
				}
				if follow == nil || len(g.To(immedDom)) > len(g.To(follow)) {
					follow = immedDom
				}
			}
			head.SwitchFollow = follow
			// Tag nodes that belong to the switch.
			switchNodes[head] = true
			head.SwitchHead = head
			traversed := make(map[*cfg.Node]bool)
			for _, headSucc := range headSuccs {
				flagSwitchNodes(g, dom, head, follow, headSucc, switchNodes, traversed)
			}
			if follow != nil {
				follow.SwitchHead = head
			}
			prim := &primitive.Switch{
				Head: head.DOTID(),
			}
			if follow != nil {
				prim.Follow = follow.DOTID()
			}
			var ns []graph.Node
			for n := range switchNodes {
				ns = append(ns, n)
			}
			for _, n := range cfg.SortByRevPost(ns) {
				prim.Nodes = append(prim.Nodes, n.DOTID())
			}
			pretty.Println("switch:", prim)
		}
	}
}

// --- [ tagNodesInCase  ] -----------------------------------------------------

// flagSwitchNodes recursively tags nodes that belong to the switch described by
// the map switchNodes, and the header and follow node of the switch.
func flagSwitchNodes(g *cfg.Graph, dom path.DominatorTree, head, follow, n *cfg.Node, switchNodes map[*cfg.Node]bool, traversed map[*cfg.Node]bool) {
	traversed[n] = true
	if n == follow {
		return
	}
	if len(g.From(n)) > 2 {
		return
	}
	if immedDom := node(dom.DominatorOf(n)); !switchNodes[immedDom] {
		return
	}
	switchNodes[n] = true
	n.SwitchHead = head
	for _, succ := range cfg.SortByRevPost(g.From(n)) {
		if traversed[succ] {
			continue
		}
		flagSwitchNodes(g, dom, head, follow, succ, switchNodes, traversed)
	}
}

// isSuccessor reports whether n is a successor of head.
func isSuccessor(g *cfg.Graph, n, head graph.Node) bool {
	for _, succ := range g.From(head) {
		if n == succ {
			return true
		}
	}
	return false
}

// --- [ structLoops ] ---------------------------------------------------------

// structLoop structures loops in the given control flow graph.
func structLoop(g *cfg.Graph, prims *primitive.Primitives) {
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
				// Check that the latching node is at the same nesting level of case
				// statement (if any).
				dbg.Println("located latch node:", latch.DOTID())
				if latch.SwitchHead != nil && latch.SwitchHead == I.h.SwitchHead {
					fmt.Println("   SKIP: latch switch head:", latch.SwitchHead.DOTID())
					fmt.Println("   SKIP: interval head switch head:", I.h.SwitchHead.DOTID())
					continue
				}
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
func findNodesInLoop(g *cfg.Graph, I *Interval, latch *cfg.Node, dom path.DominatorTree) *primitive.Loop {
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
	loop := &primitive.Loop{
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

// --- [ structIf ] ------------------------------------------------------------

// structIf structures if-statements in the given control flow graph.
//
// Pre-condition: the nodes of the graph are numbered in reverse post-order.
func structIf(g *cfg.Graph, prims *primitive.Primitives, dom path.DominatorTree) {
	// TODO: Ensure that the header and latch nodes of loops are correctly
	// labelled, so they are not considered if-statements. It is possible, quite
	// likely even, that the current code updates n.LoopHeader for the inveral
	// nodes, (but not the corresponding nodes in the original graph).

	unresolved := make(map[*cfg.Node]bool)

	// TODO: Verify that reverse dfsLast order = reverse post-order.

	// Reverse reverse post-order (from Figure 13, 1993)
	for _, n := range cfg.SortByPost(g.Nodes()) {
		// TODO: Check if !(loop header) should be determined with
		//    n.LoopHead != n
		// rather than
		//    n.LoopHead == nil

		// 2-way condition, not loop header, not loop latch
		if len(g.From(n)) == 2 && n.LoopHead == nil {
			// possible follow node.
			var follow *cfg.Node
			followInEdges := 0
			// find all nodes that have this node as immediate dominator.
			for _, m := range dom.DominatedBy(n) {
				mm := node(m)
				nInEdges := len(g.To(mm))
				// TODO: calculate on the fly instead of relying on isBackEdge calculation.
				nBackEdges := mm.NBackEdges
				if nInEdges-nBackEdges > followInEdges {
					follow = mm
					followInEdges = nInEdges - nBackEdges
				}
			}
			if follow != nil && followInEdges > 1 {
				dbg.Printf("follow of %v: %v\n", n.DOTID(), follow.DOTID())
				prim := &primitive.If{
					Cond:   n.DOTID(),
					Follow: follow.DOTID(),
				}
				n.IfFollow = follow
				// Assign the follow node to all unresolved nodes.
				for m := range unresolved {
					m.IfFollow = follow
					delete(unresolved, m)
					// TODO: Figure out the purpose of unresolved. For what type of
					// CFGs do they appear?
					prim.Unresolved = append(prim.Unresolved, m.DOTID())
				}
				prims.Ifs = append(prims.Ifs, prim)
			} else {
				unresolved[n] = true
			}
		}
	}
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
