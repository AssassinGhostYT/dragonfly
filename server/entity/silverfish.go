package entity

import (
	mobsx "github.com/AssassinGhostYT/MobsX-MC"
	"github.com/AssassinGhostYT/MobsX-MC/behavior"
	"github.com/AssassinGhostYT/MobsX-MC/sensor"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"math/rand"
)

type Silverfish struct {
	brain     *mobsx.Brain
	navigator *mobsx.Navigator
	mc        *MovementComputer
	self      *Ent

	health  float64
	alerted *behavior.CallForHelpBehavior
}

func NewSilverfish(opts world.EntitySpawnOpts) *world.EntityHandle {
	s := &Silverfish{health: 8}
	return opts.New(SilverfishType, s)
}

func (s *Silverfish) Apply(data *world.EntityData) {
	s.mc = &MovementComputer{Gravity: 0.08, Drag: 0.02, StepHeight: 1.0}
	data.Data = s
}

func (s *Silverfish) Tick(e *Ent, tx *world.Tx) *Movement {
	s.self = e
	if s.Dead() {
		return nil
	}

	if s.brain == nil {
		s.brain = mobsx.NewBrain()
		wBridge := WorldBridge{E: e}
		s.navigator = mobsx.NewNavigator(EntityBridge{E: e, tx: tx}, wBridge)
		s.navigator.Speed = 0.25

		playerScanner := &sensor.PlayerSensor{Range: 32}
		s.alerted = &behavior.CallForHelpBehavior{RangeX: 21, RangeY: 11, RangeZ: 21}

		attack := behavior.NewAttack(playerScanner, s.navigator)
		attack.AttackRange = 1.1

		s.brain.AddSensor(playerScanner)
		s.brain.AddBehavior(s.alerted)
		s.brain.AddBehavior(attack)
		s.brain.AddBehavior(&behavior.InfestBehavior{InfestChance: 0.1})
		s.brain.AddBehavior(behavior.NewWander(s.navigator, 10))
	}

	wBridge := WorldBridge{E: e}
	s.navigator.Sync(wBridge)
	s.brain.Tick(EntityBridge{E: e, tx: tx}, wBridge)

	pos := cube.PosFromVec3(e.Position())
	b := tx.Block(pos)

	if n, ok := b.(interface {
		EncodeBlock() (string, map[string]any)
	}); ok {
		name, _ := n.EncodeBlock()
		if name == "minecraft:lava" || name == "minecraft:flowing_lava" {
			s.Hurt(4.0, block.FireDamageSource{})
		}
	}

	if len(b.Model().BBox(pos, tx)) > 0 {
		s.Hurt(1.0, SuffocationDamageSource{})
	}

	for player := range tx.Players() {
		if p, ok := player.(interface {
			GameMode() world.GameMode
			Hurt(damage float64, src world.DamageSource) (n float64, v bool)
			Position() mgl64.Vec3
		}); ok {
			if p.GameMode().AllowsTakingDamage() {
				dist := p.Position().Sub(e.Position()).Len()
				if dist < 1.1 {
					dmg := 1.0
					if tx.World().Difficulty() == world.DifficultyHard {
						dmg = 1.5
					}
					p.Hurt(dmg, AttackDamageSource{Attacker: e})
				}
			}
		}
	}

	if e.Velocity().LenSqr() > 0.001 && rand.Intn(10) == 0 {
		tx.PlaySound(e.Position(), sound.SilverfishStep{})
	}
	if rand.Intn(100) == 0 {
		tx.PlaySound(e.Position(), sound.SilverfishAmbient{})
	}

	m := s.mc.TickMovement(e, e.data.Pos, e.data.Vel, e.data.Rot, tx)
	e.data.Pos, e.data.Vel = m.pos, m.vel
	return m
}

var SilverfishType silverfishType

type silverfishType struct{}

func (t silverfishType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	e := &Ent{tx: tx, handle: handle, data: data}
	if s, ok := data.Data.(*Silverfish); ok {
		s.self = e
	}
	return e
}
func (silverfishType) EncodeEntity() string { return "minecraft:silverfish" }

func (silverfishType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.23, 0, -0.23, 0.23, 0.32, 0.23)
}
func (silverfishType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	s := &Silverfish{health: 8}
	s.Apply(data)
}
func (silverfishType) EncodeNBT(*world.EntityData) map[string]any { return nil }

func (s *Silverfish) Health() float64        { return s.health }
func (s *Silverfish) MaxHealth() float64     { return 8 }
func (s *Silverfish) SetMaxHealth(v float64) { s.health = v }
func (s *Silverfish) Dead() bool             { return s.health <= 0 }
func (s *Silverfish) Hurt(damage float64, src world.DamageSource) (n float64, v bool) {
	if s.Dead() {
		return 0, false
	}
	s.health -= damage
	if s.health > 0 && damage > 0 {
		if s.alerted != nil {
			s.alerted.Alerted = true
		}
		if s.self != nil {
			for _, v := range s.self.tx.Viewers(s.self.Position()) {
				v.ViewEntityAction(s.self, HurtAction{})
			}
		}
	}
	if s.health <= 0 && s.self != nil {
		s.self.tx.AddParticle(s.self.Position(), particle.Evaporate{})
		for _, v := range s.self.tx.Viewers(s.self.Position()) {
			v.ViewEntityAction(s.self, DeathAction{})
		}
		for _, handle := range NewExperienceOrbs(s.self.Position(), 5) {
			s.self.tx.AddEntity(handle)
		}
		_ = s.self.Close()
	}
	return damage, true
}
func (s *Silverfish) Heal(health float64, src world.HealingSource) { s.health += health }
func (s *Silverfish) KnockBack(src mgl64.Vec3, f, h float64) {
	if s.self == nil {
		return
	}
	s.self.data.Vel = s.mc.KnockBack(src, f, h, s.self.data.Pos)
}
func (s *Silverfish) Velocity() mgl64.Vec3     { return s.self.data.Vel }
func (s *Silverfish) SetVelocity(v mgl64.Vec3) { s.self.data.Vel = v }
func (s *Silverfish) Speed() float64           { return 0.25 }
func (s *Silverfish) SetSpeed(v float64)       {}
func (s *Silverfish) AddEffect(e effect.Effect) {}
func (s *Silverfish) RemoveEffect(e effect.Type) {}
func (s *Silverfish) Effects() []effect.Effect { return nil }
func (s *Silverfish) PistonImmovable() bool    { return false }
func (s *Silverfish) PistonBreakable() bool    { return false }

// UUID returns the unique identifier of the entity.
func (s *Silverfish) UUID() uuid.UUID {
	if s.self == nil {
		return uuid.UUID{}
	}
	return s.self.H().UUID()
}

// DeathPosition returns the death position, dimension and whether the entity died.
func (s *Silverfish) DeathPosition() (mgl64.Vec3, world.Dimension, bool) {
	if s.self == nil {
		return mgl64.Vec3{}, nil, s.Dead()
	}
	return s.self.Position(), s.self.tx.World().Dimension(), s.Dead()
}
