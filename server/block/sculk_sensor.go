package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math"
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
	switch s.Phase {
	case 1: // Active -> Transition to Cooldown
		s.Phase = 2
		s.Power = 0
		tx.SetBlock(pos, s, nil)
		tx.PlaySound(pos.Vec3Centre(), sound.SculkSensorPowerOff{})
		tx.ScheduleBlockUpdate(pos, s, 500*time.Millisecond) // 10 ticks cooldown
	case 2: // Cooldown -> Transition to Inactive
		s.Phase = 0
		tx.SetBlock(pos, s, nil)
		tx.ScheduleBlockUpdate(pos, s, 100*time.Millisecond) // Restart scanning
	case 0: // Inactive -> Scanning
		if s.scan(tx, pos) {
			return
		}
		tx.ScheduleBlockUpdate(pos, s, 100*time.Millisecond)
	}
}

// scan returns true if a vibration was detected.
func (s SculkSensor) scan(tx *world.Tx, pos cube.Pos) bool {
	// Radius of 8 blocks (Wiki: 8 block spherical radius)
	for e := range tx.EntitiesWithin(pos.Box().Grow(8)) {
		if sneak, ok := e.(interface{ Sneaking() bool }); ok && sneak.Sneaking() {
			continue
		}
		
		dist := pos.Vec3Centre().Sub(e.Position()).Len()
		if dist <= 8 {
			s.detect(tx, pos, e.Position(), e)
			return true
		}
	}
	return false
}

// EntityInside ...
func (s SculkSensor) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	s.detect(tx, pos, e.Position(), e)
}

// EntityStepOn ...
func (s SculkSensor) EntityStepOn(pos cube.Pos, tx *world.Tx, e world.Entity) {
	s.detect(tx, pos, e.Position(), e)
}

// RandomTick ...
func (s SculkSensor) RandomTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	// Kickstart the loop if it's not running
	tx.ScheduleBlockUpdate(pos, s, 100*time.Millisecond)
}

// detect attempts to detect a vibration.
func (s SculkSensor) detect(tx *world.Tx, pos cube.Pos, origin mgl64.Vec3, e world.Entity) {
	if s.Phase != 0 {
		return
	}

	dist := pos.Vec3Centre().Sub(origin).Len()
	if dist > 8 {
		return
	}

	s.Power = int(math.Max(1, 15-math.Floor(15.0/8.0*dist)))
	s.Phase = 1
	tx.SetBlock(pos, s, nil)

	tx.PlaySound(pos.Vec3Centre(), sound.SculkSensorPowerOn{})
	// Particle at the player/source position
	tx.AddParticle(origin, particle.VibrationSignal{Origin: origin})

	if _, ok := e.(interface{ GameMode() world.GameMode }); ok {
		s.activateNearbyShriekers(tx, pos)
	}

	// Stay on for 30 ticks (1.5 seconds)
	tx.ScheduleBlockUpdate(pos, s, 1500*time.Millisecond)
}

// activateNearbyShriekers searches for and activates shriekers within 8 blocks.
func (s SculkSensor) activateNearbyShriekers(tx *world.Tx, pos cube.Pos) {
	for x := -8; x <= 8; x++ {
		for y := -8; y <= 8; y++ {
			for z := -8; z <= 8; z++ {
				p := pos.Add(cube.Pos{x, y, z})
				if shrieker, ok := tx.Block(p).(SculkShrieker); ok {
					if shrieker.CanSummon {
						shrieker.shriek(tx, p, 0)
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
	tx.ScheduleBlockUpdate(pos, s, 100*time.Millisecond)
	return placed(ctx)
}

// Source ...
func (SculkSensor) Source() bool { return true }

// RedstoneSource ...
func (SculkSensor) RedstoneSource() bool { return true }

// WeakPower ...
func (s SculkSensor) WeakPower(cube.Pos, cube.Face, *world.Tx, bool) int { return s.Power }

// StrongPower ...
func (s SculkSensor) StrongPower(cube.Pos, cube.Face, *world.Tx, bool) int { return 0 }

// LightEmissionLevel ...
func (s SculkSensor) LightEmissionLevel() uint8 {
	if s.Phase == 1 { return 1 }
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
