package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// receivedPower returns the highest level of redstone power received by the block at the position passed.
func receivedPower(pos cube.Pos, tx *world.Tx) int {
	var power int
	for _, face := range cube.Faces() {
		sidePos := pos.Side(face)
		b := tx.Block(sidePos)
		if c, ok := b.(world.Conductor); ok {
			p := c.WeakPower(sidePos, face.Opposite(), tx)
			if p > power {
				power = p
			}
		}
		// Redstone dust and other sources might provide power through blocks.
		// For now, we only check direct neighbors as a base.
	}
	return power
}

// RedstoneBlock is a block that emits a constant redstone signal.
type RedstoneBlock struct {
	solid
}

// Source ...
func (RedstoneBlock) Source() bool {
	return true
}

// WeakPower ...
func (RedstoneBlock) WeakPower(cube.Pos, cube.Face, *world.Tx) int {
	return 15
}

// StrongPower ...
func (RedstoneBlock) StrongPower(cube.Pos, cube.Face, *world.Tx) int {
	return 0
}

// EncodeItem ...
func (RedstoneBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_block", 0
}

// EncodeBlock ...
func (RedstoneBlock) EncodeBlock() (string, map[string]any) {
	return "minecraft:redstone_block", nil
}
