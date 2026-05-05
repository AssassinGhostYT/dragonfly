package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// RedstoneDust is a block that transmits redstone power.
type RedstoneDust struct {
	transparent
	empty

	// Power is the current redstone power level (0-15).
	Power int
}

// BreakInfo ...
func (r RedstoneDust) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(item.Redstone{}))
}

// EncodeItem ...
func (r RedstoneDust) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_wire", 0
}

// EncodeBlock ...
func (r RedstoneDust) EncodeBlock() (string, map[string]any) {
	return "minecraft:redstone_wire", map[string]any{"redstone_signal": int32(r.Power)}
}

// UseOnBlock ...
func (r RedstoneDust) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, r)
	if !used {
		return false
	}
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceUp, tx) {
		return false
	}

	r.Power = r.calculatePower(pos, tx)
	place(tx, pos, r, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (r RedstoneDust) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceUp, tx) {
		breakBlock(r, pos, tx)
		return
	}
	
	newPower := r.calculatePower(pos, tx)
	if newPower != r.Power {
		r.Power = newPower
		tx.SetBlock(pos, r, nil)
	}
}

// WeakPower ...
func (r RedstoneDust) WeakPower(pos cube.Pos, face cube.Face, tx *world.Tx) int {
	if face == cube.FaceUp {
		return r.Power
	}
	if face == cube.FaceDown {
		return 0
	}
	return r.Power
}

// StrongPower ...
func (r RedstoneDust) StrongPower(pos cube.Pos, face cube.Face, tx *world.Tx) int {
	return r.WeakPower(pos, face, tx)
}

// calculatePower returns the highest level of received redstone power at the provided position.
func (r RedstoneDust) calculatePower(pos cube.Pos, tx *world.Tx) int {
	var maxPower int
	for _, face := range cube.Faces() {
		sidePos := pos.Side(face)
		b := tx.Block(sidePos)
		
		// Check for direct sources
		if c, ok := b.(world.Conductor); ok {
			p := c.WeakPower(sidePos, face.Opposite(), tx)
			if p > maxPower {
				maxPower = p
			}
		}

		// Check for wire propagation
		if wire, ok := b.(RedstoneDust); ok {
			if wire.Power-1 > maxPower {
				maxPower = wire.Power - 1
			}
		}
		
		// Check for power through blocks (only if the neighbor is solid)
		if _, ok := b.Model().(model.Solid); ok {
			// Redstone propagation through blocks could be implemented here.
		}
	}
	return maxPower
}

// allRedstoneDust ...
func allRedstoneDust() (dust []world.Block) {
	for i := 0; i < 16; i++ {
		dust = append(dust, RedstoneDust{Power: i})
	}
	return
}
