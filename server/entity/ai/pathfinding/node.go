package pathfinding

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

// Node represents a single point in the pathfinding grid.
type Node struct {
	Pos    cube.Pos
	Parent *Node
	G, H   float64
}

// F returns the total cost of the node (G + H).
func (n *Node) F() float64 {
	return n.G + n.H
}

// Path represents a calculated sequence of positions for an entity to follow.
type Path struct {
	Nodes []cube.Pos
	Index int
}

// Next returns the next position in the path and advances the index.
func (p *Path) Next() (cube.Pos, bool) {
	if p.Index >= len(p.Nodes) {
		return cube.Pos{}, false
	}
	pos := p.Nodes[p.Index]
	p.Index++
	return pos, true
}

// AtEnd returns true if the path has been fully traversed.
func (p *Path) AtEnd() bool {
	return p.Index >= len(p.Nodes)
}

// Distance calculates the Euclidean distance between two positions.
func Distance(a, b cube.Pos) float64 {
	return mgl64.Vec3{float64(a.X()), float64(a.Y()), float64(a.Z())}.Sub(mgl64.Vec3{float64(b.X()), float64(b.Y()), float64(b.Z())}).Len()
}
