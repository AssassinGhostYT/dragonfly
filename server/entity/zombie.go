package entity

import (
	"github.com/AssassinGhostYT/MobsX-MC"
	"github.com/AssassinGhostYT/MobsX-MC/behavior"
	"github.com/AssassinGhostYT/MobsX-MC/sensor"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"math/rand"
	"time"
)

// Zombie is a common hostile mob that spawns in dark areas.
type Zombie struct {
	brain     *mobsx.Brain
	navigator *mobsx.Navigator
	mc        *MovementComputer
	self      *Ent

	health float64
}

// NewZombie creates a new Zombie entity.
func NewZombie(opts world.EntitySpawnOpts) *world.EntityHandle {
	z := &Zombie{health: 20}
	return opts.New(ZombieType, z)
}

func (z *Zombie) Apply(data *world.EntityData) {
	z.mc = &MovementComputer{Gravity: 0.08, Drag: 0.02, StepHeight: 0.6}
	data.Data = z
}

func (z *Zombie) Tick(e *Ent, tx *world.Tx) *Movement {
	z.self = e
	if z.Dead() {
		return nil
	}

	if z.brain == nil {
		z.brain = mobsx.NewBrain()
		wBridge := WorldBridge{E: e}
		z.navigator = mobsx.NewNavigator(EntityBridge{E: e, tx: tx}, wBridge)
		z.navigator.Speed = 0.23

		playerScanner := &sensor.PlayerSensor{Range: 16}
		attack := behavior.NewAttack(playerScanner, z.navigator)
		attack.Damage = 3.0
		attack.Cooldown = time.Second * 2

		z.brain.AddSensor(playerScanner)
		z.brain.AddBehavior(attack)
		z.brain.AddBehavior(behavior.NewWander(z.navigator, 10))
	}

	wBridge := WorldBridge{E: e}
	z.navigator.Sync(wBridge)
	z.brain.Tick(EntityBridge{E: e, tx: tx}, wBridge)

	// Zombie burning in sun logic
	pos := cube.PosFromVec3(e.Position())
	if tx.World().Time()%24000 < 12000 && tx.World().SkyLight(pos) > 10 && tx.HighestBlock(pos.X(), pos.Z()) <= pos.Y() {
		e.OnFire()
	}

	if rand.Intn(100) == 0 {
		tx.PlaySound(e.Position(), sound.ZombieAmbient{})
	}

	m := z.mc.TickMovement(e, e.data.Pos, e.data.Vel, e.data.Rot, tx)
	e.data.Pos, e.data.Vel = m.pos, m.vel
	return m
}

var ZombieType zombieType

type zombieType struct{}

func (t zombieType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	e := &Ent{tx: tx, handle: handle, data: data}
	if z, ok := data.Data.(*Zombie); ok {
		z.self = e
	}
	return e
}
func (zombieType) EncodeEntity() string { return "minecraft:zombie" }

func (zombieType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.3, 0, -0.3, 0.3, 1.9, 0.3)
}
func (zombieType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	z := &Zombie{health: 20}
	z.Apply(data)
}
func (zombieType) EncodeNBT(*world.EntityData) map[string]any { return nil }

func (z *Zombie) Health() float64        { return z.health }
func (z *Zombie) MaxHealth() float64     { return 20 }
func (z *Zombie) SetMaxHealth(v float64) { z.health = v }
func (z *Zombie) Dead() bool             { return z.health <= 0 }
func (z *Zombie) Hurt(damage float64, src world.DamageSource) (n float64, v bool) {
	if z.Dead() {
		return 0, false
	}
	z.health -= damage
	if z.health > 0 && damage > 0 {
		z.self.tx.PlaySound(z.self.Position(), sound.ZombieHurt{})
		for _, v := range z.self.tx.Viewers(z.self.Position()) {
			v.ViewEntityAction(z.self, HurtAction{})
		}
	}
	if z.health <= 0 && z.self != nil {
		z.self.tx.PlaySound(z.self.Position(), sound.ZombieDeath{})
		z.self.tx.AddParticle(z.self.Position(), particle.Evaporate{})
		for _, v := range z.self.tx.Viewers(z.self.Position()) {
			v.ViewEntityAction(z.self, DeathAction{})
		}
		// XP drops
		for _, handle := range NewExperienceOrbs(z.self.Position(), 5) {
			z.self.tx.AddEntity(handle)
		}
		_ = z.self.Close()
	}
	return damage, true
}
func (z *Zombie) Heal(health float64, src world.HealingSource) { z.health += health }
func (z *Zombie) KnockBack(src mgl64.Vec3, f, h float64) {
	if z.self == nil {
		return
	}
	z.self.data.Vel = z.mc.KnockBack(src, f, h, z.self.data.Pos)
}
func (z *Zombie) Velocity() mgl64.Vec3     { return z.self.data.Vel }
func (z *Zombie) SetVelocity(v mgl64.Vec3) { z.self.data.Vel = v }
func (z *Zombie) Speed() float64           { return 0.23 }
func (z *Zombie) SetSpeed(v float64)       {}
func (z *Zombie) AddEffect(e effect.Effect) {}
func (z *Zombie) RemoveEffect(e effect.Type) {}
func (z *Zombie) Effects() []effect.Effect { return nil }
func (z *Zombie) PistonImmovable() bool    { return false }
func (z *Zombie) PistonBreakable() bool    { return false }

func (z *Zombie) UUID() uuid.UUID {
	if z.self == nil {
		return uuid.UUID{}
	}
	return z.self.H().UUID()
}

func (z *Zombie) DeathPosition() (mgl64.Vec3, world.Dimension, bool) {
	if z.self == nil {
		return mgl64.Vec3{}, nil, z.Dead()
	}
	return z.self.Position(), z.self.tx.World().Dimension(), z.Dead()
}
