package interval

import "github.com/graphism/exp/cfg"

// Primitives records the control flow primitives of a function.
type Primitives struct {
	// TODO: include information about the nodes contained within each interval.
	// Loops.
	Loops []*Loop `json:"loops"`
	// TODO: Add if-statements.
}

// A Loop is a loop control flow primitive.
type Loop struct {
	// Loop type.
	Type cfg.LoopType
	// Header of the loop.
	Head string
	// Latch node of the loop.
	Latch string
	// Follow node of the loop.
	Follow string
	// Nodes of the loop.
	Nodes []string
}
