package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SculkCatalyst is a block that spreads sculk when a mob dies near it.
type SculkCatalyst struct {
	solid
	// Bloom specifies if the sculk catalyst is currently blooming.
	Bloom bool
}

// UseOnBlock ...
func (c SculkCatalyst) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, ok := firstReplaceable(tx, pos, face, c)
	if !ok {
		return false
	}

	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

// LightEmissionLevel ...
func (SculkCatalyst) LightEmissionLevel() uint8 {
	return 6
}

// BreakInfo ...
func (s SculkCatalyst) BreakInfo() BreakInfo {
	return newBreakInfo(3, alwaysHarvestable, hoeEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(s, 1)}
		}
		return nil
	}).withXPDropRange(20, 20).withBlastResistance(3)
}

// EncodeItem ...
func (SculkCatalyst) EncodeItem() (name string, meta int16) {
	return "minecraft:sculk_catalyst", 0
}

// EncodeBlock ...
func (c SculkCatalyst) EncodeBlock() (string, map[string]any) {
	return "minecraft:sculk_catalyst", map[string]any{"bloom": boolByte(c.Bloom)}
}

// allSculkCatalysts ...
func allSculkCatalysts() (b []world.Block) {
	return []world.Block{SculkCatalyst{Bloom: true}, SculkCatalyst{Bloom: false}}
}
