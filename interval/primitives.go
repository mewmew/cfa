package interval

import "github.com/graphism/exp/cfg"

// Primitives records the control flow primitives of a function.
type Primitives struct {
	// map from collapsed node name to the nodes of the corresponding interval.
	Intervals map[string][]string `json:"intervals"`
	// TODO: include information about the nodes contained within each interval.
	// Loops.
	Loops []*Loop `json:"loops"`
	// TODO: Add if-statements.
}

// NewPrimitives returns a new record for the control flow primitives of a
// function.
func NewPrimitives() *Primitives {
	return &Primitives{
		Intervals: make(map[string][]string),
	}
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
