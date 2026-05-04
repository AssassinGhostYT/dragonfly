package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// RedstoneTorch is a non-solid block that emits redstone power.
type RedstoneTorch struct {
	transparent
	empty

	// Facing is the direction from the torch to the block it is attached to.
	Facing cube.Face
	// Lit is true if the torch is lit.
	Lit bool
}

// BreakInfo ...
func (t RedstoneTorch) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(RedstoneTorch{Lit: true}))
}

// LightEmissionLevel ...
func (t RedstoneTorch) LightEmissionLevel() uint8 {
	if t.Lit {
		return 7
	}
	return 0
}

// WeakPower ...
func (t RedstoneTorch) WeakPower(pos cube.Pos, face cube.Face, tx *world.Tx) int {
	if t.Lit && t.Facing != face {
		return 15
	}
	return 0
}

// StrongPower ...
func (t RedstoneTorch) StrongPower(pos cube.Pos, face cube.Face, tx *world.Tx) int {
	if t.Lit && t.Facing == cube.FaceDown && face == cube.FaceUp {
		return 15
	}
	return 0
}

// UseOnBlock ...
func (t RedstoneTorch) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, t)
	if !used {
		return false
	}
	if face == cube.FaceDown {
		return false
	}
	if !tx.Block(pos.Side(face.Opposite())).Model().FaceSolid(pos.Side(face.Opposite()), face, tx) {
		found := false
		for _, i := range []cube.Face{cube.FaceSouth, cube.FaceWest, cube.FaceNorth, cube.FaceEast, cube.FaceDown} {
			if tx.Block(pos.Side(i)).Model().FaceSolid(pos.Side(i), i.Opposite(), tx) {
				found = true
				face = i.Opposite()
				break
			}
		}
		if !found {
			return false
		}
	}
	t.Facing = face.Opposite()
	t.Lit = receivedPower(pos.Side(t.Facing), tx) == 0

	place(tx, pos, t, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (t RedstoneTorch) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(t.Facing)).Model().FaceSolid(pos.Side(t.Facing), t.Facing.Opposite(), tx) {
		breakBlock(t, pos, tx)
		return
	}
	lit := receivedPower(pos.Side(t.Facing), tx) == 0
	if lit != t.Lit {
		t.Lit = lit
		tx.SetBlock(pos, t, nil)
	}
}

// HasLiquidDrops ...
func (t RedstoneTorch) HasLiquidDrops() bool {
	return true
}

// EncodeItem ...
func (t RedstoneTorch) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_torch", 0
}

// EncodeBlock ...
func (t RedstoneTorch) EncodeBlock() (name string, properties map[string]any) {
	var face string
	switch t.Facing {
	case cube.FaceDown:
		face = "top"
	case unknownFace:
		face = "unknown"
	default:
		face = t.Facing.String()
	}

	name = "minecraft:redstone_torch"
	if !t.Lit {
		name = "minecraft:unlit_redstone_torch"
	}
	return name, map[string]any{"torch_facing_direction": face}
}

// allRedstoneTorches ...
func allRedstoneTorches() (torch []world.Block) {
	for _, face := range cube.Faces() {
		if face == cube.FaceUp {
			face = unknownFace
		}
		torch = append(torch, RedstoneTorch{Facing: face, Lit: true})
		torch = append(torch, RedstoneTorch{Facing: face, Lit: false})
	}
	return
}
