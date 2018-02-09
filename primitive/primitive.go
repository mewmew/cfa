package primitive

import "github.com/graphism/exp/cfg"

// Primitives records the control flow primitives of a function.
type Primitives struct {
	// map from collapsed node name to the nodes of the corresponding interval.
	Intervals map[string][]string `json:"intervals"`
	// TODO: include information about the nodes contained within each interval.
	// Loops.
	Loops []*Loop `json:"loops"`
	// If-statements.
	Ifs []*If `json:"ifs"`
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

// An If is 2-way conditional control flow primitive.
type If struct {
	// Conditional node.
	Cond string
	// Follow node of the 2-way conditional.
	Follow string
	// Unresolved nodes of the if-statement.
	Unresolved []string
}
