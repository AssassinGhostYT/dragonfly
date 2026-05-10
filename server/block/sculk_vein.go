package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SculkVein is a block that spreads sculk.
type SculkVein struct {
	replaceable
	transparent
	empty
	sourceWaterDisplacer

	// North, South, West, East, Up, Down specify if the sculk vein is attached to that face.
	North, South, West, East, Up, Down bool
}

// WithAttachment returns a SculkVein with an attachment on the given face.
func (v SculkVein) WithAttachment(face cube.Face, attached bool) SculkVein {
	switch face {
	case cube.FaceDown:
		v.Down = attached
	case cube.FaceUp:
		v.Up = attached
	case cube.FaceSouth:
		v.South = attached
	case cube.FaceWest:
		v.West = attached
	case cube.FaceNorth:
		v.North = attached
	case cube.FaceEast:
		v.East = attached
	}
	return v
}

// UseOnBlock ...
func (v SculkVein) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, ok := firstReplaceable(tx, pos, face, v)
	if !ok {
		return false
	}

	if b, ok := tx.Block(pos).(SculkVein); ok {
		v = b
	}
	v = v.WithAttachment(face.Opposite(), true)

	ctx.IgnoreEntityCollision()
	place(tx, pos, v, user, ctx)
	return placed(ctx)
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
	if v.Down {
		bits |= 1
	}
	if v.Up {
		bits |= 2
	}
	if v.South {
		bits |= 4
	}
	if v.West {
		bits |= 8
	}
	if v.North {
		bits |= 16
	}
	if v.East {
		bits |= 32
	}
	return "minecraft:sculk_vein", map[string]any{"multi_face_direction_bits": bits}
}

// allSculkVeins ...
func allSculkVeins() (b []world.Block) {
	for i := 0; i < 64; i++ {
		b = append(b, SculkVein{
			Down:  i&1 != 0,
			Up:    i&2 != 0,
			South: i&4 != 0,
			West:  i&8 != 0,
			North: i&16 != 0,
			East:  i&32 != 0,
		})
	}
	return
}
