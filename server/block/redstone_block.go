package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// RedstoneBlock is a solid block that emits a full redstone signal.
type RedstoneBlock struct {
	solid
}

// WeakPower ...
func (RedstoneBlock) WeakPower(cube.Pos, cube.Face, *world.Tx) int {
	return 15
}

// StrongPower ...
func (RedstoneBlock) StrongPower(cube.Pos, cube.Face, *world.Tx) int {
	return 0
}

// BreakInfo ...
func (r RedstoneBlock) BreakInfo() BreakInfo {
	return newBreakInfo(5, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierWood.HarvestLevel
	}, pickaxeEffective, oneOf(r)).withBlastResistance(30)
}

// EncodeItem ...
func (RedstoneBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_block", 0
}

// EncodeBlock ...
func (RedstoneBlock) EncodeBlock() (string, map[string]any) {
	return "minecraft:redstone_block", nil
}
