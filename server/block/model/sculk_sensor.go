package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// SculkSensor is the model used by sculk sensors. It has a base with a flat top surface
// that entities can stand on, triggering EntityStepOn.
type SculkSensor struct{}

// BBox returns a BBox for the sculk sensor base.
func (SculkSensor) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	const h = 8.0 / 16
	return []cube.BBox{
		cube.Box(0, 0, 0, 1, h, 1),
	}
}

// FaceSolid returns true for the bottom face.
func (SculkSensor) FaceSolid(_ cube.Pos, face cube.Face, _ world.BlockSource) bool {
	return face == cube.FaceDown
}
