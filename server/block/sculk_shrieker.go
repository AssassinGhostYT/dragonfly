package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// SculkShrieker is a block that shrieks and can summon a warden.
type SculkShrieker struct {
	empty
	transparent
	sourceWaterDisplacer
	// CanSummon specifies if the sculk shrieker can summon a warden.
	CanSummon bool
	// Shrieking specifies if the sculk shrieker is currently shrieking.
	Shrieking bool
}

// BreakInfo ...
func (s SculkShrieker) BreakInfo() BreakInfo {
	return newBreakInfo(3, alwaysHarvestable, hoeEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(s, 1)}
		}
		return nil
	}).withXPDropRange(5, 5).withBlastResistance(3)
}

// EncodeItem ...
func (SculkShrieker) EncodeItem() (name string, meta int16) {
	return "minecraft:sculk_shrieker", 0
}

// EncodeBlock ...
func (s SculkShrieker) EncodeBlock() (string, map[string]any) {
	return "minecraft:sculk_shrieker", map[string]any{
		"can_summon": boolByte(s.CanSummon),
		"active":     boolByte(s.Shrieking),
	}
}

// allSculkShriekers ...
func allSculkShriekers() (b []world.Block) {
	for _, c := range []bool{true, false} {
		for _, s := range []bool{true, false} {
			b = append(b, SculkShrieker{CanSummon: c, Shrieking: s})
		}
	}
	return
}
