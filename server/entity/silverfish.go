package entity

import (
	"github.com/AssassinGhostYT/MobsX-MC"
	"github.com/AssassinGhostYT/MobsX-MC/behavior"
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
}

// NewSilverfish creates a new Silverfish entity handle.
func NewSilverfish(opts world.EntitySpawnOpts) *world.EntityHandle {
	return opts.New(SilverfishType, Silverfish{})
}

// Apply ...
func (s Silverfish) Apply(data *world.EntityData) {
	s.mc = &MovementComputer{Gravity: 0.08, Drag: 0.02}
	data.Data = &s
}

// Tick ...
func (s *Silverfish) Tick(e *Ent, tx *world.Tx) *Movement {
	if s.brain == nil {
		s.brain = mobsx.NewBrain()
		wBridge := worldBridge{tx: tx}
		s.navigator = mobsx.NewNavigator(EntityBridge{E: e}, wBridge)
		s.navigator.Speed = 0.15
		s.brain.AddBehavior(behavior.NewWander(s.navigator, 10))
	}
	
	s.brain.Tick(EntityBridge{E: e}, worldBridge{tx: tx})
	
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
	return cube.NewBBox(-0.15, 0, -0.15, 0.15, 0.3, 0.15)
}
func (silverfishType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	Silverfish{}.Apply(data)
}
func (silverfishType) EncodeNBT(*world.EntityData) map[string]any { return nil }
