package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
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

	place(tx, pos, r, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (r RedstoneDust) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceUp, tx) {
		breakBlock(r, pos, tx)
		return
	}
	// TODO: Implement complex redstone wire power propagation logic.
}

// WeakPower ...
func (r RedstoneDust) WeakPower(cube.Pos, cube.Face, *world.Tx) int {
	return r.Power
}

// StrongPower ...
func (r RedstoneDust) StrongPower(cube.Pos, cube.Face, *world.Tx) int {
	return 0
}

// allRedstoneDust ...
func allRedstoneDust() (dust []world.Block) {
	for i := 0; i < 16; i++ {
		dust = append(dust, RedstoneDust{Power: i})
	}
	return
}
