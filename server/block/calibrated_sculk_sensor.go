package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// CalibratedSculkSensor is a block that detects vibrations at a specific frequency.
type CalibratedSculkSensor struct {
	empty
	transparent
	sourceWaterDisplacer
	// Phase is the current phase of the sculk sensor. 0 is inactive, 1 is active, 2 is cooldown.
	Phase int
	// Facing is the direction the sculk sensor is facing.
	Facing cube.Face
}

// UseOnBlock ...
func (c CalibratedSculkSensor) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, ok := firstReplaceable(tx, pos, face, c)
	if !ok {
		return false
	}
	c.Facing = user.Rotation().Direction().Face().Opposite()
	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

// Source ...
func (CalibratedSculkSensor) Source() bool {
	return true
}

// RedstoneSource ...
func (CalibratedSculkSensor) RedstoneSource() bool {
	return true
}

// WeakPower ...
func (c CalibratedSculkSensor) WeakPower(cube.Pos, cube.Face, *world.Tx, bool) int {
	// TODO: Implement power based on vibration frequency.
	return 0
}

// StrongPower ...
func (c CalibratedSculkSensor) StrongPower(cube.Pos, cube.Face, *world.Tx, bool) int {
	return 0
}

// LightEmissionLevel ...
func (c CalibratedSculkSensor) LightEmissionLevel() uint8 {
	if c.Phase == 1 {
		return 1
	}
	return 0
}

// BreakInfo ...
func (c CalibratedSculkSensor) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, alwaysHarvestable, hoeEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(c, 1)}
		}
		return nil
	}).withXPDropRange(5, 5).withBlastResistance(1.5)
}

// EncodeItem ...
func (CalibratedSculkSensor) EncodeItem() (name string, meta int16) {
	return "minecraft:calibrated_sculk_sensor", 0
}

// EncodeBlock ...
func (c CalibratedSculkSensor) EncodeBlock() (string, map[string]any) {
	return "minecraft:calibrated_sculk_sensor", map[string]any{
		"sculk_sensor_phase":           int32(c.Phase),
		"minecraft:cardinal_direction": c.Facing.String(),
	}
}

// allCalibratedSculkSensors ...
func allCalibratedSculkSensors() (b []world.Block) {
	for _, p := range []int{0, 1, 2} {
		for _, f := range []cube.Face{cube.FaceNorth, cube.FaceSouth, cube.FaceWest, cube.FaceEast} {
			b = append(b, CalibratedSculkSensor{Phase: p, Facing: f})
		}
	}
	return
}
