package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math"
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

// Model ...
func (SculkSensor) Model() world.BlockModel {
	return model.SculkSensor{}
}

// UseOnBlock ...
func (s SculkSensor) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, ok := firstReplaceable(tx, pos, face, s)
	if !ok {
		return false
	}
	place(tx, pos, s, user, ctx)
	if placed(ctx) {
		s.startScanLoop(tx.World(), pos)
		return true
	}
	return false
}

// startScanLoop begins a periodic scan for entities within 8 blocks.
func (s SculkSensor) startScanLoop(w *world.World, pos cube.Pos) {
	time.AfterFunc(100*time.Millisecond, func() {
		w.Exec(func(tx *world.Tx) {
			b := tx.Block(pos)
			sensor, ok := b.(SculkSensor)
			if !ok {
				return
			}
			if sensor.Phase == 0 {
				detected := sensor.runScan(tx, pos)
				if !detected {
					// No detection, schedule next scan
					sensor.startScanLoop(w, pos)
				}
				// If detected, the cooldown timer in detect() will restart the loop
			}
		})
	})
}

// runScan checks for entities within 8 blocks and detects the closest one.
// Returns true if an entity was detected.
func (s SculkSensor) runScan(tx *world.Tx, pos cube.Pos) bool {
	var closest mgl64.Vec3
	var closestEntity world.Entity
	closestDist := math.MaxFloat64
	centre := pos.Vec3Centre()
	box := cube.Box(
		float64(pos.X()-8), float64(pos.Y()-8), float64(pos.Z()-8),
		float64(pos.X()+8), float64(pos.Y()+8), float64(pos.Z()+8),
	)

	for e := range tx.EntitiesWithin(box) {
		s.checkEntity(e, centre, &closest, &closestEntity, &closestDist)
	}
	for p := range tx.Players() {
		s.checkEntity(p, centre, &closest, &closestEntity, &closestDist)
	}
	if closestEntity != nil {
		s.detect(tx, pos, closest, closestEntity)
		return true
	}
	return false
}

func (s SculkSensor) checkEntity(e world.Entity, centre mgl64.Vec3, closest *mgl64.Vec3, closestEntity *world.Entity, closestDist *float64) {
	if sneak, ok := e.(interface{ Sneaking() bool }); ok && sneak.Sneaking() {
		return
	}
	origin := e.Position()
	dist := centre.Sub(origin).Len()
	if dist <= 8 && dist < *closestDist {
		*closest = origin
		*closestEntity = e
		*closestDist = dist
	}
}

// EntityInside ...
func (s SculkSensor) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	s.detect(tx, pos, e.Position(), e)
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

	dist := pos.Vec3Centre().Sub(origin).Len()
	if dist > 8 {
		return
	}

	s.Power = int(math.Max(1, 15-math.Floor(15.0/8.0*dist)))
	s.Phase = 1
	tx.SetBlock(pos, s, nil)

	tx.PlaySound(pos.Vec3Centre(), sound.SculkSensorPowerOn{})
	tx.AddParticle(pos.Vec3Centre(), particle.VibrationSignal{Origin: origin})

	if _, ok := e.(interface{ GameMode() world.GameMode }); ok {
		s.activateNearbyShriekers(tx, pos)
	}

	w := tx.World()
	time.AfterFunc(1500*time.Millisecond, func() {
		w.Exec(func(tx *world.Tx) {
			b := tx.Block(pos)
			if sensor, ok := b.(SculkSensor); ok && sensor.Phase == 1 {
				sensor.Phase = 2
				sensor.Power = 0
				tx.SetBlock(pos, sensor, nil)
				tx.PlaySound(pos.Vec3Centre(), sound.SculkSensorPowerOff{})

				time.AfterFunc(500*time.Millisecond, func() {
					w.Exec(func(tx *world.Tx) {
						b := tx.Block(pos)
						if sensor, ok := b.(SculkSensor); ok && sensor.Phase == 2 {
							sensor.Phase = 0
							tx.SetBlock(pos, sensor, nil)
							sensor.startScanLoop(w, pos)
						}
					})
				})
			}
		})
	})
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

// compile-time check: ensure ScheduledTick is not needed (removed to avoid hash-mismatch issues)
var _ world.Block = SculkSensor{}
