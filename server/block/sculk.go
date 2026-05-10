package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Sculk is a block found in the deep dark biome.
type Sculk struct {
	solid
}

// UseOnBlock ...
func (s Sculk) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, ok := firstReplaceable(tx, pos, face, s)
	if !ok {
		return false
	}

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (s Sculk) BreakInfo() BreakInfo {
	return newBreakInfo(0.6, alwaysHarvestable, hoeEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(s, 1)}
		}
		return nil
	}).withXPDropRange(1, 1).withBlastResistance(0.6)
}

// EncodeItem ...
func (Sculk) EncodeItem() (name string, meta int16) {
	return "minecraft:sculk", 0
}

// EncodeBlock ...
func (Sculk) EncodeBlock() (string, map[string]any) {
	return "minecraft:sculk", map[string]any{}
}
