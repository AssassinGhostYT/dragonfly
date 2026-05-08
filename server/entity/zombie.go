package entity

import (
	mobsx "github.com/AssassinGhostYT/MobsX-MC"
	"github.com/AssassinGhostYT/MobsX-MC/behavior"
	"github.com/AssassinGhostYT/MobsX-MC/sensor"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
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
	attack    *behavior.AttackBehavior
	scanner   *sensor.PlayerSensor

	health    float64
	deadTicks int
}

// NewZombie creates a new Zombie entity.
func NewZombie(opts world.EntitySpawnOpts) *world.EntityHandle {
	z := &Zombie{health: 20}
	return opts.New(ZombieType, z)
}

func (z *Zombie) Apply(data *world.EntityData) {
	z.mc = &MovementComputer{Gravity: 0.08, Drag: 0.02, StepHeight: 1.0}
	data.Data = z
}

func (z *Zombie) Tick(e *Ent, tx *world.Tx) *Movement {
	z.self = e
	if z.Dead() {
		z.deadTicks++
		if z.deadTicks >= 20 {
			_ = e.Close()
		}
		return nil
	}

	if z.brain == nil {
		z.brain = mobsx.NewBrain()
		wBridge := WorldBridge{E: e}
		z.navigator = mobsx.NewNavigator(EntityBridge{E: e, tx: tx}, wBridge)
		z.navigator.Speed = 0.23

		z.scanner = &sensor.PlayerSensor{Range: 35} // Follow range 35
		z.attack = behavior.NewAttack(z.scanner, z.navigator)

		z.brain.AddSensor(z.scanner)
		z.brain.AddBehavior(z.attack)
		z.brain.AddBehavior(behavior.NewWander(z.navigator, 80)) // Even slower wander
	}

	// Update damage based on difficulty
	diff := tx.World().Difficulty()
	switch diff {
	case world.DifficultyEasy:
		z.attack.Damage = 2.5
	case world.DifficultyNormal:
		z.attack.Damage = 3.0
	case world.DifficultyHard:
		z.attack.Damage = 4.5
	default:
		z.attack.Damage = 3.0
	}
	z.attack.Cooldown = time.Second * 2

	// Adjust speed based on whether it has a target
	if len(z.scanner.Detected) > 0 {
		z.navigator.Speed = 0.23 // Standard chase speed
	} else {
		z.navigator.Speed = 0.12 // Slower wander speed
	}

	wBridge := WorldBridge{E: e}
	z.navigator.Sync(wBridge)
	z.brain.Tick(EntityBridge{E: e, tx: tx}, wBridge)

	// If not moving and no target, look around randomly
	if z.navigator.Path.AtEnd() && len(z.scanner.Detected) == 0 {
		if rand.Intn(20) == 0 {
			curYaw, curPitch := e.Rotation().Yaw(), e.Rotation().Pitch()
			newYaw := curYaw + (rand.Float64()*60 - 30)
			newPitch := curPitch + (rand.Float64()*20 - 10)
			if newPitch > 20 {
				newPitch = 20
			} else if newPitch < -20 {
				newPitch = -20
			}
			e.data.Rot = cube.Rotation{newYaw, newPitch}
		}
	}

	// Zombie burning in sun logic
	pos := cube.PosFromVec3(e.Position())
	if tx.World().Time()%24000 < 12000 && tx.SkyLight(pos) > 10 && tx.HighestBlock(pos.X(), pos.Z()) <= pos.Y() {
		e.SetOnFire(time.Second * 8)
	}

	if e.OnFireDuration() > 0 && tx.World().Time()%20 == 0 {
		z.Hurt(1.0, block.FireDamageSource{})
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
	// 2 armor points (8% reduction)
	damage *= 0.92

	z.health -= damage
	if z.health > 0 && damage > 0 {
		z.self.tx.PlaySound(z.self.Position(), sound.ZombieHurt{})
		for _, v := range z.self.tx.Viewers(z.self.Position()) {
			v.ViewEntityAction(z.self, HurtAction{})
		}
	}
	if z.health <= 0 && z.self != nil {
		z.self.tx.PlaySound(z.self.Position(), sound.ZombieDeath{})
		for _, v := range z.self.tx.Viewers(z.self.Position()) {
			v.ViewEntityAction(z.self, DeathAction{})
		}
		// XP drops
		for _, handle := range NewExperienceOrbs(z.self.Position(), 5) {
			z.self.tx.AddEntity(handle)
		}
		// Item drops
		for _, it := range z.Drops() {
			z.self.tx.AddEntity(NewItem(world.EntitySpawnOpts{Position: z.self.Position()}, it))
		}
	}
	return damage, true
}
func (z *Zombie) Heal(health float64, src world.HealingSource) { z.health += health }
func (z *Zombie) KnockBack(src mgl64.Vec3, f, h float64) {
	if z.self == nil {
		return
	}
	// 5% Knockback resistance
	f *= 0.95
	h *= 0.95
	z.self.data.Vel = z.mc.KnockBack(src, f, h, z.self.data.Pos)
}
func (z *Zombie) Velocity() mgl64.Vec3       { return z.self.data.Vel }
func (z *Zombie) SetVelocity(v mgl64.Vec3)   { z.self.data.Vel = v }
func (z *Zombie) Speed() float64             { return 0.23 }
func (z *Zombie) SetSpeed(v float64)         {}
func (z *Zombie) AddEffect(e effect.Effect)  {}
func (z *Zombie) RemoveEffect(e effect.Type) {}
func (z *Zombie) Effects() []effect.Effect   { return nil }
func (z *Zombie) PistonImmovable() bool      { return false }
func (z *Zombie) PistonBreakable() bool      { return false }

// Drops returns the drops of the zombie.
func (z *Zombie) Drops() []item.Stack {
	return []item.Stack{item.NewStack(item.RottenFlesh{}, rand.Intn(3))}
}

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
