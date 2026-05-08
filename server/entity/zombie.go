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
	attack    *behavior.AttackBehavior
	scanner   *sensor.PlayerSensor

	health    float64
	deadTicks int

	baby               bool
	equipment          []item.Stack
	doorBreakingTicks  int
	currentDoorPos     cube.Pos
}

// NewZombie creates a new Zombie entity.
func NewZombie(opts world.EntitySpawnOpts) *world.EntityHandle {
	z := &Zombie{health: 20}
	if rand.Intn(20) == 0 { // 5% chance of being a baby
		z.baby = true
	}
	return opts.New(ZombieType, z)
}

// NewZombieBaby creates a new Baby Zombie entity.
func NewZombieBaby(opts world.EntitySpawnOpts) *world.EntityHandle {
	z := &Zombie{health: 20, baby: true}
	return opts.New(ZombieType, z)
}

func (z *Zombie) Apply(data *world.EntityData) {
	z.mc = &MovementComputer{Gravity: 0.08, Drag: 0.02, StepHeight: 1.0}
	if z.baby {
		z.mc.StepHeight = 0.5 // Can pass 1x1 gaps
	}
	data.Data = z

	// Natural equipment
	if !z.baby {
		if rand.Intn(100) < 5 { // 5% chance of equipment
			r := rand.Intn(3)
			switch r {
			case 0:
				z.equipment = append(z.equipment, item.NewStack(item.Shovel{Tier: item.ToolTierIron}, 1))
			case 1:
				z.equipment = append(z.equipment, item.NewStack(item.Sword{Tier: item.ToolTierIron}, 1))
			case 2:
				if rand.Intn(100) < 1 {
					z.equipment = append(z.equipment, item.NewStack(item.Sword{Tier: item.ToolTierDiamond}, 1))
				}
			}
		}
	}
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
		z.navigator.Finder.Height = 2
		if z.baby {
			z.navigator.Finder.Height = 1
		}
		z.navigator.Speed = 0.23
		if z.baby {
			z.navigator.Speed = 0.32 // 30% faster
		}

		z.scanner = &sensor.PlayerSensor{Range: 35}
		z.attack = behavior.NewAttack(z.scanner, z.navigator)
		z.attack.AttackRange = 1.3

		z.brain.AddSensor(z.scanner)
		z.brain.AddBehavior(z.attack)
		z.brain.AddBehavior(behavior.NewWander(z.navigator, 80))
	}

	// Wiki Damage Scaling
	diff := tx.World().Difficulty()
	baseDmg := 3.0 // Normal
	if diff == world.DifficultyEasy {
		baseDmg = 2.5
	} else if diff == world.DifficultyHard {
		baseDmg = 4.5
	}

	weaponDmg := 0.0
	if len(z.equipment) > 0 {
		if w, ok := z.equipment[0].Item().(item.Weapon); ok {
			weaponDmg = w.AttackDamage() - 1 // Subtract player base 1 dmg
		}
	}

	totalDmg := baseDmg + weaponDmg
	if weaponDmg > 0 {
		// Weapon scaling
		if diff == world.DifficultyEasy {
			totalDmg = 0.5*totalDmg + 1
		} else if diff == world.DifficultyHard {
			totalDmg = 1.5 * totalDmg
		}
	}
	z.attack.Damage = totalDmg
	z.attack.Cooldown = time.Second * 2

	// Speed Adjustment
	speed := 0.23
	wanderSpeed := 0.12
	if z.baby {
		speed = 0.45
		wanderSpeed = 0.22
	}
	if len(z.scanner.Detected) > 0 {
		z.navigator.Speed = speed
	} else {
		z.navigator.Speed = wanderSpeed
	}

	// Adult pathfinding is now handled by Finder.Height = 2
	// But we keep this as backup for immediate reaction
	if !z.baby {
	headPos := cube.PosFromVec3(e.Position().Add(mgl64.Vec3{0, 1.8, 0}))
	if tx.Block(headPos).Model().FaceSolid(headPos, cube.FaceDown, tx) {
		z.navigator.Speed = 0
	}
	}
	wBridge := WorldBridge{E: e}
	z.navigator.Sync(wBridge)
	z.brain.Tick(EntityBridge{E: e, tx: tx}, wBridge)

	// Door breaking logic (Wiki: ~10 seconds = 200 ticks)
	if !z.baby && diff == world.DifficultyHard {
		// Priority: Check for door in front
		var doorPos cube.Pos
		rot := e.Rotation()
		lookingAt := cube.PosFromVec3(e.Position().Add(cube.Rotation{rot.Yaw(), 0}.Vec3().Mul(0.7)))
		
		if door, ok := tx.Block(lookingAt).(block.WoodDoor); ok && !door.Open {
			doorPos = lookingAt
		}

		if doorPos != (cube.Pos{}) {
			if doorPos != z.currentDoorPos {
				if z.currentDoorPos != (cube.Pos{}) {
					z.stopCracking(tx, z.currentDoorPos)
				}
				z.currentDoorPos = doorPos
				z.doorBreakingTicks = 0
				z.startCracking(tx, doorPos)
			}
			z.doorBreakingTicks++
			e.data.Vel = mgl64.Vec3{} // Stop moving while breaking
			if z.doorBreakingTicks%20 == 0 {
				tx.PlaySound(doorPos.Vec3(), sound.ZombieAmbient{})
				z.updateCracking(tx, doorPos, z.doorBreakingTicks)
			}
			if z.doorBreakingTicks >= 200 {
				tx.SetBlock(doorPos, block.Air{}, nil)
				tx.PlaySound(doorPos.Vec3(), sound.DoorCrash{})
				z.doorBreakingTicks = 0
				z.currentDoorPos = cube.Pos{}
			}
			return m
		} else if z.currentDoorPos != (cube.Pos{}) {
			z.stopCracking(tx, z.currentDoorPos)
			z.currentDoorPos = cube.Pos{}
			z.doorBreakingTicks = 0
		}
	}

	// Idle rotation
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

	// Burning logic
	pos := cube.PosFromVec3(e.Position())
	if tx.World().Time()%24000 > 23460 || tx.World().Time()%24000 < 12000 { // 27s before day
		if tx.SkyLight(pos) > 14 && tx.HighestBlock(pos.X(), pos.Z()) <= pos.Y() {
			// Wiki: check water/helmet etc (simplified for now)
			e.SetOnFire(time.Second * 8)
		}
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

func (zombieType) BBox(e world.Entity) cube.BBox {
	if ent, ok := e.(*Ent); ok {
		if z, ok := ent.data.Data.(*Zombie); ok && z.baby {
			return cube.Box(-0.245, 0, -0.245, 0.245, 0.98, 0.245)
		}
	}
	return cube.Box(-0.3, 0, -0.3, 0.3, 1.9, 0.3)
}
func (zombieType) DecodeNBT(m map[string]any, data *world.EntityData) {
	z := &Zombie{health: 20}
	if baby, ok := m["IsBaby"].(byte); ok && baby == 1 {
		z.baby = true
	}
	z.Apply(data)
}
func (zombieType) EncodeNBT(data *world.EntityData) map[string]any {
	if z, ok := data.Data.(*Zombie); ok {
		baby := byte(0)
		if z.baby {
			baby = 1
		}
		return map[string]any{"IsBaby": baby}
	}
	return nil
}

func (z *Zombie) Health() float64        { return z.health }
func (z *Zombie) MaxHealth() float64     { return 20 }
func (z *Zombie) SetMaxHealth(v float64) { z.health = v }
func (z *Zombie) Dead() bool             { return z.health <= 0 }
func (z *Zombie) Baby() bool             { return z.baby }
func (z *Zombie) Scale() float64 {
	if z.baby {
		return 0.5
	}
	return 1.0
}
func (z *Zombie) InteractText() string { return "Generar Zombie Bebé" }
func (z *Zombie) Experience() int {
	if z.baby {
		return 12
	}
	return 5
}
func (z *Zombie) Hurt(damage float64, src world.DamageSource) (n float64, v bool) {
	if z.Dead() {
		return 0, false
	}
	damage *= 0.92 // 2 armor points
	z.health -= damage
	if z.health > 0 && damage > 0 {
		z.self.tx.PlaySound(z.self.Position(), sound.ZombieHurt{})
		for _, v := range z.self.tx.Viewers(z.self.Position()) {
			v.ViewEntityAction(z.self, HurtAction{})
		}
	}
	if z.health <= 0 && z.self != nil {
		z.self.tx.AddParticle(z.self.Position(), particle.Evaporate{})
		z.self.tx.PlaySound(z.self.Position(), sound.ZombieDeath{})
		for _, v := range z.self.tx.Viewers(z.self.Position()) {
			v.ViewEntityAction(z.self, DeathAction{})
		}
		
		// Spread drops to avoid shadows and ensure they can be picked up
		spawnPos := z.self.Position().Add(mgl64.Vec3{0, 0.5, 0})
		for _, handle := range NewExperienceOrbs(spawnPos, z.Experience()) {
			z.self.tx.AddEntity(handle)
		}
		for _, it := range z.Drops() {
			if it.Count() > 0 {
				opts := world.EntitySpawnOpts{
					Position: spawnPos,
					Velocity: mgl64.Vec3{rand.Float64()*0.1 - 0.05, 0.2, rand.Float64()*0.1 - 0.05},
				}
				z.self.tx.AddEntity(NewItem(opts, it))
			}
		}
	}
	return damage, true
}
func (z *Zombie) Heal(health float64, src world.HealingSource) { z.health += health }
func (z *Zombie) KnockBack(src mgl64.Vec3, f, h float64) {
	if z.self == nil {
		return
	}
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
	drops := []item.Stack{item.NewStack(item.RottenFlesh{}, rand.Intn(3))}
	for _, it := range z.equipment {
		if rand.Intn(100) < 15 {
			drops = append(drops, it)
		}
	}
	return drops
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
