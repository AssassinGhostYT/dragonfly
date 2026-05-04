package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
)

// pistonFace converts a cube.Face to the specific integer ID used by pistons in this version of Bedrock.
// According to the Wiki: 0:Up, 1:Down, 2:South, 3:North, 4:East, 5:West.
func pistonFace(f cube.Face) int32 {
	switch f {
	case cube.FaceUp:
		return 0
	case cube.FaceDown:
		return 1
	case cube.FaceSouth:
		return 2
	case cube.FaceNorth:
		return 3
	case cube.FaceEast:
		return 4
	case cube.FaceWest:
		return 5
	}
	return 0
}

// faceFromPiston converts the piston's integer ID back to a cube.Face.
func faceFromPiston(i int32) cube.Face {
	switch i {
	case 0:
		return cube.FaceUp
	case 1:
		return cube.FaceDown
	case 2:
		return cube.FaceSouth
	case 3:
		return cube.FaceNorth
	case 4:
		return cube.FaceEast
	case 5:
		return cube.FaceWest
	}
	return cube.FaceDown
}
