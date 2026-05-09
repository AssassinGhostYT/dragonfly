package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// SculkCatalyst is a block that spreads sculk when a mob dies near it.
type SculkCatalyst struct {
	solid
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
func (SculkCatalyst) EncodeBlock() (string, map[string]any) {
	return "minecraft:sculk_catalyst", nil
}
