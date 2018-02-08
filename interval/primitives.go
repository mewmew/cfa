package interval

import "github.com/graphism/exp/cfg"

// A Primitive is a control flow primitive.
type Primitive interface {
	// isPrim ensures that only control flow primitives may be assigned to
	// Primitive.
	isPrim()
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

// isPrim ensures that only control flow primitives may be assigned to
// Primitive.
func (*Loop) isPrim() {}
