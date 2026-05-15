package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
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

// ScheduledTick ...
func (s SculkSensor) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if s.Phase == 1 {
		// Active -> Cooldown (40 ticks later)
		s.Phase = 2
		s.Power = 0
		tx.SetBlock(pos, s, nil)
		tx.ScheduleBlockUpdate(pos, s, time.Millisecond*50) // 1 tick cooldown
	} else if s.Phase == 2 {
		// Cooldown -> Inactive (1 tick later)
		s.Phase = 0
		tx.SetBlock(pos, s, nil)
	}
}

// EntityStepOn ...
func (s SculkSensor) EntityStepOn(pos cube.Pos, tx *world.Tx, e world.Entity) {
	s.detect(tx, pos, e.Position(), e)
}

// detect attempts to detect a vibration.
func (s SculkSensor) detect(tx *world.Tx, pos cube.Pos, origin mgl64.Vec3, e world.Entity) {
	if s.Phase != 0 {
		return
	}

	// Wiki: "Agacharse impide la creación de vibraciones al moverse"
	if sneak, ok := e.(interface{ Sneaking() bool }); ok && sneak.Sneaking() {
		return
	}

	// Calculate redstone power (inversely proportional to distance)
	dist := pos.Vec3Centre().Sub(origin).Len()
	if dist > 8 {
		return
	}

	// TODO: Wool occlusion (RayTrace)

	s.Phase = 1
	s.Power = int(max(1, 15-(dist*15/8)))
	tx.SetBlock(pos, s, nil)

	// Emit vibration signal particle from origin to sensor
	tx.AddParticle(pos.Vec3Centre(), particle.VibrationSignal{Origin: origin})

	// Wiki: "Las vibraciones causadas por el jugador también pueden activar a los chilladores sculk cercanos."
	if _, ok := e.(interface{ GameMode() world.GameMode }); ok {
		s.activateNearbyShriekers(tx, pos)
	}

	// Schedule transition to cooldown after 40 ticks (2 seconds)
	tx.ScheduleBlockUpdate(pos, s, 2*time.Second)
}

// activateNearbyShriekers searches for and activates shriekers within 8 blocks.
func (s SculkSensor) activateNearbyShriekers(tx *world.Tx, pos cube.Pos) {
	for x := -8; x <= 8; x++ {
		for y := -8; y <= 8; y++ {
			for z := -8; z <= 8; z++ {
				p := pos.Add(cube.Pos{x, y, z})
				if shrieker, ok := tx.Block(p).(SculkShrieker); ok {
					// Only naturally generated shriekers or those that can summon
					if shrieker.CanSummon {
						shrieker.shriek(tx, p, 0) // TODO: Get actual player warning level
					}
				}
			}
		}
	}
}

// UseOnBlock ...
func (s SculkSensor) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, ok := firstReplaceable(tx, pos, face, s)
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

// RedstoneSource ...
func (SculkSensor) RedstoneSource() bool {
	return true
}

// WeakPower ...
func (s SculkSensor) WeakPower(cube.Pos, cube.Face, *world.Tx, bool) int {
	return s.Power
}

// StrongPower ...
func (s SculkSensor) StrongPower(cube.Pos, cube.Face, *world.Tx, bool) int {
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
