package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// RedstoneTorch is a non-solid blocks that emits little light and also a full-strength redstone signal when lit.
type RedstoneTorch struct {
	transparent
	empty

	// Facing is the direction from the torch to the block.
	Facing cube.Face
	// Lit is if the redstone torch is lit and emitting power.
	Lit bool
}

// LightEmissionLevel ...
func (t RedstoneTorch) LightEmissionLevel() uint8 {
	if t.Lit {
		return 7
	}
	return 0
}

// BreakInfo ...
func (t RedstoneTorch) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(t))
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

	t.Facing = face.Opposite()
	t.Lit = t.isLit(pos, tx)

	place(tx, pos, t, user, ctx)
	return placed(ctx)
}

// isLit checks if the torch should be lit based on the power received by its support block.
func (t RedstoneTorch) isLit(pos cube.Pos, tx *world.Tx) bool {
	supportPos := pos.Side(t.Facing)
	support := tx.Block(supportPos)
	if _, ok := support.(solid); ok {
		// Zig-zag logic: Redstone torches turn off if their support block is powered.
		return receivedPower(supportPos, tx) == 0
	}
	return true
}

// NeighbourUpdateTick ...
func (t RedstoneTorch) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(t.Facing)).Model().FaceSolid(pos.Side(t.Facing), t.Facing.Opposite(), tx) {
		breakBlock(t, pos, tx)
		return
	}
	lit := t.isLit(pos, tx)
	if lit != t.Lit {
		t.Lit = lit
		tx.SetBlock(pos, t, nil)
	}
}

// Source ...
func (t RedstoneTorch) Source() bool {
	return t.Lit
}

// WeakPower ...
func (t RedstoneTorch) WeakPower(_ cube.Pos, face cube.Face, _ *world.Tx) int {
	if !t.Lit {
		return 0
	}
	// Redstone torches power all blocks except the one they are attached to.
	if face == t.Facing {
		return 0
	}
	return 15
}

// StrongPower ...
func (t RedstoneTorch) StrongPower(_ cube.Pos, face cube.Face, _ *world.Tx) int {
	if t.Lit && face == cube.FaceDown {
		return 15
	}
	return 0
}

// EncodeItem ...
func (t RedstoneTorch) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_torch", 0
}

// EncodeBlock ...
func (t RedstoneTorch) EncodeBlock() (string, map[string]any) {
	face := "unknown"
	if t.Facing != unknownFace {
		face = t.Facing.String()
		if t.Facing == cube.FaceDown {
			face = "top" // Attached to the floor
		}
	}
	if t.Lit {
		return "minecraft:redstone_torch", map[string]any{"torch_facing_direction": face}
	}
	return "minecraft:unlit_redstone_torch", map[string]any{"torch_facing_direction": face}
}

// Hash ...
func (t RedstoneTorch) Hash() (uint64, uint64) {
	return hashRedstoneTorch, uint64(pistonFace(t.Facing)) | uint64(boolByte(t.Lit))<<3
}

// allRedstoneTorches ...
func allRedstoneTorches() (all []world.Block) {
	for _, f := range append(cube.Faces(), unknownFace) {
		if f == cube.FaceUp {
			continue
		}
		all = append(all, RedstoneTorch{Facing: f, Lit: true})
		all = append(all, RedstoneTorch{Facing: f, Lit: false})
	}
	return
}
