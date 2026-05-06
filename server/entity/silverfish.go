package entity

import (
	"github.com/AssassinGhostYT/MobsX-MC"
	"github.com/AssassinGhostYT/MobsX-MC/behavior"
	"github.com/AssassinGhostYT/MobsX-MC/sensor"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// Silverfish is a small insect-like hostile mob.
type Silverfish struct {
	brain     *mobsx.Brain
	navigator *mobsx.Navigator
	mc        *MovementComputer

	health  float64
	alerted *behavior.CallForHelpBehavior
}

// NewSilverfish creates a new Silverfish entity handle.
func NewSilverfish(opts world.EntitySpawnOpts) *world.EntityHandle {
	return opts.New(SilverfishType, Silverfish{health: 8})
}

// Apply ...
func (s Silverfish) Apply(data *world.EntityData) {
	s.mc = &MovementComputer{Gravity: 0.08, Drag: 0.02}
	data.Data = &s
}

// Tick se ejecuta en cada tic del servidor para actualizar la IA y el movimiento.
func (s *Silverfish) Tick(e *Ent, tx *world.Tx) *Movement {
	if s.brain == nil {
		s.brain = mobsx.NewBrain()
		// Usamos el puntero a la entidad 'e' para que los puentes siempre tengan acceso al tx actual.
		eBridge := EntityBridge{E: e}
		wBridge := worldBridge{E: e}

		s.navigator = mobsx.NewNavigator(eBridge, wBridge)
		s.navigator.Speed = 0.2

		playerScanner := &sensor.PlayerSensor{Range: 16}
		s.alerted = &behavior.CallForHelpBehavior{RangeX: 21, RangeY: 11, RangeZ: 21}

		s.brain.AddSensor(playerScanner)
		s.brain.AddBehavior(s.alerted)
		s.brain.AddBehavior(behavior.NewAttack(playerScanner, s.navigator))
		s.brain.AddBehavior(&behavior.InfestBehavior{InfestChance: 0.01})
		s.brain.AddBehavior(behavior.NewWander(s.navigator, 10))
	}

	wBridge := worldBridge{E: e}
	s.navigator.Sync(wBridge)
	s.brain.Tick(EntityBridge{E: e}, wBridge)

	if rand.Intn(100) == 0 {
		tx.PlaySound(e.Position(), sound.SilverfishAmbient{})
	}

	m := s.mc.TickMovement(e, e.data.Pos, e.data.Vel, e.data.Rot, tx)
	e.data.Pos, e.data.Vel = m.pos, m.vel
	return m
}

// SilverfishType ...
var SilverfishType silverfishType

type silverfishType struct{}

func (t silverfishType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}
func (silverfishType) EncodeEntity() string { return "minecraft:silverfish" }
func (silverfishType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.15, 0, -0.15, 0.15, 0.3, 0.15)
}
func (silverfishType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	Silverfish{health: 8}.Apply(data)
}
func (silverfishType) EncodeNBT(*world.EntityData) map[string]any { return nil }

// Living methods
func (s *Silverfish) Health() float64        { return s.health }
func (s *Silverfish) MaxHealth() float64     { return 8 }
func (s *Silverfish) SetMaxHealth(v float64) { s.health = v }
func (s *Silverfish) Dead() bool             { return s.health <= 0 }
func (s *Silverfish) Hurt(damage float64, src world.DamageSource) (n float64, v bool) {
	s.health -= damage
	if s.alerted != nil {
		s.alerted.Alerted = true
	}
	return damage, true
}
func (s *Silverfish) Heal(health float64, src world.HealingSource) { s.health += health }
func (s *Silverfish) KnockBack(src mgl64.Vec3, f, h float64)       {}
func (s *Silverfish) Velocity() mgl64.Vec3                         { return mgl64.Vec3{} }
func (s *Silverfish) SetVelocity(v mgl64.Vec3)                     {}
func (s *Silverfish) Speed() float64                               { return 0.2 }
func (s *Silverfish) SetSpeed(v float64)                           {}
func (s *Silverfish) AddEffect(e any)                              {}
func (s *Silverfish) RemoveEffect(e any)                           {}
func (s *Silverfish) Effects() []any                               { return nil }
func (s *Silverfish) PistonImmovable() bool                        { return false }
func (s *Silverfish) PistonBreakable() bool                        { return false }
