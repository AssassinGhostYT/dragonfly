package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
)

// SculkVein is a block that spreads sculk.
type SculkVein struct {
	replaceable
	transparent
	empty
	sourceWaterDisplacer

	// Directions is a slice of all directions the sculk vein is attached to.
	Directions []cube.Face
}

// BreakInfo ...
func (v SculkVein) BreakInfo() BreakInfo {
	return newBreakInfo(0.2, alwaysHarvestable, hoeEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(v, 1)}
		}
		return nil
	}).withXPDropRange(1, 1).withBlastResistance(0.2)
}

// EncodeItem ...
func (SculkVein) EncodeItem() (name string, meta int16) {
	return "minecraft:sculk_vein", 0
}

// EncodeBlock ...
func (v SculkVein) EncodeBlock() (string, map[string]any) {
	var bits int32
	for _, f := range v.Directions {
		switch f {
		case cube.FaceDown:
			bits |= 1
		case cube.FaceUp:
			bits |= 2
		case cube.FaceSouth:
			bits |= 4
		case cube.FaceWest:
			bits |= 8
		case cube.FaceNorth:
			bits |= 16
		case cube.FaceEast:
			bits |= 32
		}
	}
	return "minecraft:sculk_vein", map[string]any{"multi_face_direction_bits": bits}
}

// allSculkVeins ...
func allSculkVeins() (b []world.Block) {
	b = append(b, SculkVein{})
	for _, f := range cube.Faces() {
		b = append(b, SculkVein{Directions: []cube.Face{f}})
	}
	return
}
