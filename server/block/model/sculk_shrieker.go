package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// SculkShrieker is the model used by sculk shriekers.
type SculkShrieker struct{}

// BBox returns a BBox with a frame around the edges and a flat surface in the center.
func (SculkShrieker) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	const w = 2.0 / 16
	const h = 8.0 / 16
	return []cube.BBox{
		// North panel
		cube.Box(0, 0, 0, 1, 1, w),
		// South panel
		cube.Box(0, 0, 1-w, 1, 1, 1),
		// West panel
		cube.Box(0, 0, w, w, 1, 1-w),
		// East panel
		cube.Box(1-w, 0, w, 1, 1, 1-w),
		// Center surface
		cube.Box(w, h, w, 1-w, h+1.0/16, 1-w),
	}
}

// FaceSolid returns true for all faces except up.
func (SculkShrieker) FaceSolid(_ cube.Pos, face cube.Face, _ world.BlockSource) bool {
	return face != cube.FaceUp
}
