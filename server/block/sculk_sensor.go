package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SculkSensor is a block that detects vibrations.
type SculkSensor struct {
	empty
	transparent
	sourceWaterDisplacer
	// Phase is the current phase of the sculk sensor. 0 is inactive, 1 is active, 2 is cooldown.
	Phase int
	// Power is the redstone power level emitted by the sensor.
	Power int
}

// UseOnBlock ...
func (s SculkSensor) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, ok := firstReplaceable(tx, pos, face, s)
	if !ok {
		return false
	}

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// Source ...
func (SculkSensor) Source() bool {
	return true
}

// WeakPower ...
func (s SculkSensor) WeakPower(cube.Pos, cube.Face, *world.Tx) int {
	return s.Power
}

// StrongPower ...
func (s SculkSensor) StrongPower(cube.Pos, cube.Face, *world.Tx) int {
	return 0
}

// LightEmissionLevel ...
func (s SculkSensor) LightEmissionLevel() uint8 {
	if s.Phase == 1 {
		return 1
	}
	return 0
}

// BreakInfo ...
func (s SculkSensor) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, alwaysHarvestable, hoeEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(s, 1)}
		}
		return nil
	}).withXPDropRange(5, 5).withBlastResistance(1.5)
}

// EncodeItem ...
func (SculkSensor) EncodeItem() (name string, meta int16) {
	return "minecraft:sculk_sensor", 0
}

// EncodeBlock ...
func (s SculkSensor) EncodeBlock() (string, map[string]any) {
	return "minecraft:sculk_sensor", map[string]any{"sculk_sensor_phase": int32(s.Phase)}
}

// allSculkSensors ...
func allSculkSensors() (b []world.Block) {
	for _, p := range []int{0, 1, 2} {
		b = append(b, SculkSensor{Phase: p})
	}
	return
}
