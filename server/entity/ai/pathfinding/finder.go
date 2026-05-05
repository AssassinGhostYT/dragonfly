package pathfinding

import (
	"container/heap"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Finder implements the A* pathfinding algorithm for a voxel world.
type Finder struct {
	tx *world.Tx
}

// NewFinder creates a new A* pathfinder.
func NewFinder(tx *world.Tx) *Finder {
	return &Finder{tx: tx}
}

// FindPath calculates the shortest path between start and end.
func (f *Finder) FindPath(start, end cube.Pos) (Path, bool) {
	openSet := &priorityQueue{}
	heap.Init(openSet)
	heap.Push(openSet, &Node{Pos: start, G: 0, H: Distance(start, end)})

	closedSet := make(map[cube.Pos]struct{})

	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*Node)

		if current.Pos == end {
			return f.reconstructPath(current), true
		}

		closedSet[current.Pos] = struct{}{}

		for _, neighborPos := range f.neighbors(current.Pos) {
			if _, ok := closedSet[neighborPos]; ok {
				continue
			}

			gScore := current.G + 1 // Basic cost
			// TODO: Add block weight (lava, fall damage, etc.)

			neighborNode := &Node{
				Pos:    neighborPos,
				Parent: current,
				G:      gScore,
				H:      Distance(neighborPos, end),
			}
			
			heap.Push(openSet, neighborNode)
		}
	}

	return Path{}, false
}

// neighbors returns valid walkable neighbors for a position.
func (f *Finder) neighbors(pos cube.Pos) []cube.Pos {
	results := []cube.Pos{}
	// Horizontal and Vertical checks
	for _, face := range cube.Faces() {
		side := pos.Side(face)
		if f.isWalkable(side) {
			results = append(results, side)
		}
	}
	return results
}

// isWalkable returns true if an entity can stand on and pass through a position.
func (f *Finder) isWalkable(pos cube.Pos) bool {
	b := f.tx.Block(pos)
	// Must be air or non-solid to pass through
	if _, ok := b.(interface{ Solid() bool }); ok {
		// This is a simplification. Dragonfly uses models for collision.
		return false
	}
	// Must have a solid block below to stand on
	below := f.tx.Block(pos.Side(cube.FaceDown))
	if _, ok := below.(interface{ Solid() bool }); !ok {
		return false
	}
	return true
}

func (f *Finder) reconstructPath(endNode *Node) Path {
	nodes := []cube.Pos{}
	curr := endNode
	for curr != nil {
		nodes = append([]cube.Pos{curr.Pos}, nodes...)
		curr = curr.Parent
	}
	return Path{Nodes: nodes}
}

// priorityQueue implements heap.Interface for A* nodes.
type priorityQueue []*Node

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].F() < pq[j].F() }
func (pq priorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i] }
func (pq *priorityQueue) Push(x any)        { *pq = append(*pq, x.(*Node)) }
func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}
