package entity

import (
	mobsx "github.com/AssassinGhostYT/MobsX-MC"
	"github.com/AssassinGhostYT/MobsX-MC/behavior"
	"github.com/AssassinGhostYT/MobsX-MC/mmath"
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
)

// Chicken is a passive mob that lays eggs and is immune to fall damage.
type Chicken struct {
	brain     *mobsx.Brain
	navigator *mobsx.Navigator
	mc        *MovementComputer
	self      *Ent
	scanner   *sensor.PlayerSensor

	health float64

	baby        bool
	growthTicks int

	eggTimer int
	variant  int // 0: Tempered, 1: Warm, 2: Cold

	loveTicks     int
	breedCooldown int
	panicTicks    int
}

// NewChicken creates a new Chicken entity.
func NewChicken(opts world.EntitySpawnOpts) *world.EntityHandle {
	c := &Chicken{health: 4, eggTimer: 6000 + rand.Intn(6001)}
	if rand.Intn(20) == 0 {
		c.baby = true
		c.eggTimer = -1
	}
	return opts.New(ChickenType, c)
}

// NewChickenBaby creates a new Baby Chicken entity.
func NewChickenBaby(opts world.EntitySpawnOpts) *world.EntityHandle {
	c := &Chicken{health: 4, baby: true, eggTimer: -1}
	return opts.New(ChickenType, c)
}

func (c *Chicken) Apply(data *world.EntityData) {
	c.mc = &MovementComputer{Gravity: 0.08, Drag: 0.02, StepHeight: 0.5}
	data.Data = c
}

func (c *Chicken) Tick(e *Ent, tx *world.Tx) *Movement {
	c.self = e
	if c.Dead() {
		return nil
	}

	if c.breedCooldown > 0 {
		c.breedCooldown--
	}
	if c.panicTicks > 0 {
		c.panicTicks--
	}
	if c.loveTicks > 0 {
		c.loveTicks--
		if c.loveTicks%20 == 0 {
			for _, v := range tx.Viewers(e.Position()) {
				v.ViewEntityAction(e, InLoveAction{})
			}
		}
		
		// Search for partner
		if c.loveTicks > 0 && c.panicTicks == 0 {
			var partner *Chicken
			for other := range tx.EntitiesWithin(e.H().Type().BBox(e).Grow(8.0).Translate(e.Position())) {
				if other.H().UUID() == e.H().UUID() { continue }
				if ent, ok := other.(*Ent); ok {
					if c2, ok := ent.data.Data.(*Chicken); ok && c2.loveTicks > 0 && !c2.baby {
						partner = c2
						break
					}
				}
			}
			
			if partner != nil {
				dist := e.Position().Sub(partner.self.Position()).Len()
				if dist < 1.0 {
					// Breed!
					c.loveTicks = 0
					partner.loveTicks = 0
					c.breedCooldown = 6000
					partner.breedCooldown = 6000
					
					opts := world.EntitySpawnOpts{Position: e.Position()}
					tx.AddEntity(NewChickenBaby(opts))
					
					for _, v := range tx.Viewers(e.Position()) {
						v.ViewEntityAction(e, InLoveAction{}) // Extra hearts
					}
					for _, handle := range NewExperienceOrbs(e.Position(), rand.Intn(7)+1) {
						tx.AddEntity(handle)
					}
				} else {
					// Move to partner
					pPos := cube.PosFromVec3(partner.self.Position())
					c.navigator.SetTarget(mmath.Pos{pPos.X(), pPos.Y(), pPos.Z()})
				}
			}
		}
	}

	if c.brain == nil {
		c.brain = mobsx.NewBrain()
		wBridge := WorldBridge{E: e}
		c.navigator = mobsx.NewNavigator(EntityBridge{E: e, tx: tx}, wBridge)
		c.navigator.Finder.Height = 1
		c.navigator.Speed = 0.25

		c.scanner = &sensor.PlayerSensor{Range: 16}
		// Follow seeds (Wiki Bedrock: 16 blocks range)
		c.brain.AddBehavior(behavior.NewPanic(c.navigator))
		c.brain.AddBehavior(behavior.NewTempt(c.scanner, c.navigator, func(name string, meta int16) bool {
			return name == "minecraft:wheat_seeds" || name == "minecraft:beetroot_seeds" || name == "minecraft:melon_seeds" || name == "minecraft:pumpkin_seeds" || name == "minecraft:torchflower_seeds" || name == "minecraft:pitcher_pod" || name == "minecraft:seeds"
		}))
		if c.baby {
			c.brain.AddBehavior(behavior.NewFollowParent(c.navigator))
		}
		c.brain.AddSensor(c.scanner)
		c.brain.AddBehavior(behavior.NewWander(c.navigator, 10))
		
		// Set variant based on biome if spawned naturally
		pos := cube.PosFromVec3(e.Position())
		temp := tx.Temperature(pos)
		if temp < 0.2 {
			c.variant = 2 // Cold
		} else if temp > 1.0 {
			c.variant = 1 // Warm
		} else {
			c.variant = 0 // Tempered
		}
	}

	// Growth logic
	if c.baby {
		c.growthTicks++
		if c.growthTicks >= 24000 {
			c.grow(tx)
		}
	} else if c.eggTimer > 0 {
		c.eggTimer--
		if c.eggTimer <= 0 {
			c.layEgg(tx)
			c.eggTimer = 6000 + rand.Intn(6001)
		}
	}

	// Slow Falling (Flapping)
	if e.data.Vel[1] < -0.1 {
		e.data.Vel[1] = -0.1
	}

	wBridge := WorldBridge{E: e}
	c.navigator.Sync(wBridge)
	c.brain.Tick(EntityBridge{E: e, tx: tx}, wBridge)

	m := c.mc.TickMovement(e, e.data.Pos, e.data.Vel, e.data.Rot, tx)
	
	if rand.Intn(400) == 0 {
		tx.PlaySound(e.Position(), sound.ChickenAmbient{})
	}

	e.data.Pos, e.data.Vel = m.pos, m.vel
	return m
}

func (c *Chicken) layEgg(tx *world.Tx) {
	opts := world.EntitySpawnOpts{
		Position: c.self.Position(),
		Velocity: mgl64.Vec3{rand.Float64()*0.1 - 0.05, 0.2, rand.Float64()*0.1 - 0.05},
	}
	tx.AddEntity(NewItem(opts, item.NewStack(item.Egg{}, 1)))
	tx.PlaySound(c.self.Position(), sound.ChickenEgg{})
}

func (c *Chicken) grow(tx *world.Tx) {
	c.baby = false
	c.eggTimer = 6000 + rand.Intn(6001)
	for _, v := range tx.Viewers(c.self.Position()) {
		v.ViewEntityState(c.self)
	}
}

func (c *Chicken) Hurt(damage float64, src world.DamageSource) (n float64, v bool) {
	if c.Dead() {
		return 0, false
	}
	c.health -= damage
	c.panicTicks = 60 // 3 seconds of panic
	if c.health > 0 && damage > 0 {
		c.self.tx.PlaySound(c.self.Position(), sound.ChickenHurt{})
		for _, v := range c.self.tx.Viewers(c.self.Position()) {
			v.ViewEntityAction(c.self, HurtAction{})
		}
	}
	if c.health <= 0 && c.self != nil {
		c.self.tx.AddParticle(c.self.Position(), particle.Evaporate{})
		c.self.tx.PlaySound(c.self.Position(), sound.ChickenDeath{})
		for _, v := range c.self.tx.Viewers(c.self.Position()) {
			v.ViewEntityAction(c.self, DeathAction{})
		}
		
		spawnPos := c.self.Position().Add(mgl64.Vec3{0, 0.3, 0})
		if !c.baby {
			for _, handle := range NewExperienceOrbs(spawnPos, rand.Intn(3)+1) {
				c.self.tx.AddEntity(handle)
			}
			for _, it := range c.Drops() {
				if it.Count() > 0 {
					opts := world.EntitySpawnOpts{
						Position: spawnPos,
						Velocity: mgl64.Vec3{rand.Float64()*0.1 - 0.05, 0.2, rand.Float64()*0.1 - 0.05},
					}
					c.self.tx.AddEntity(NewItem(opts, it))
				}
			}
		}
		_ = c.self.Close()
	}
	return damage, true
}

func (c *Chicken) Drops() []item.Stack {
	var drops []item.Stack
	drops = append(drops, item.NewStack(item.Feather{}, rand.Intn(3)))
	if c.self.OnFireDuration() > 0 {
		drops = append(drops, item.NewStack(item.Chicken{Cooked: true}, 1))
	} else {
		drops = append(drops, item.NewStack(item.Chicken{}, 1))
	}
	return drops
}

var ChickenType chickenType

type chickenType struct{}

func (t chickenType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	e := &Ent{tx: tx, handle: handle, data: data}
	if c, ok := data.Data.(*Chicken); ok {
		c.self = e
	}
	return e
}
func (chickenType) EncodeEntity() string { return "minecraft:chicken" }

func (chickenType) BBox(e world.Entity) cube.BBox {
	if ent, ok := e.(*Ent); ok {
		if c, ok := ent.data.Data.(*Chicken); ok && c.baby {
			return cube.Box(-0.1, 0, -0.1, 0.1, 0.35, 0.1)
		}
	}
	return cube.Box(-0.2, 0, -0.2, 0.2, 0.7, 0.2)
}

func (chickenType) DecodeNBT(m map[string]any, data *world.EntityData) {
	c := &Chicken{health: 4}
	if baby, ok := m["IsBaby"].(byte); ok && baby == 1 {
		c.baby = true
	}
	if variant, ok := m["Variant"].(int32); ok {
		c.variant = int(variant)
	}
	c.Apply(data)
}

func (chickenType) EncodeNBT(data *world.EntityData) map[string]any {
	if c, ok := data.Data.(*Chicken); ok {
		baby := byte(0)
		if c.baby {
			baby = 1
		}
		return map[string]any{
			"IsBaby": baby,
			"Variant": int32(c.variant),
		}
	}
	return nil
}

func (c *Chicken) Health() float64        { return c.health }
func (c *Chicken) MaxHealth() float64     { return 4 }
func (c *Chicken) SetMaxHealth(v float64) { c.health = v }
func (c *Chicken) Dead() bool             { return c.health <= 0 }
func (c *Chicken) Baby() bool             { return c.baby }
func (c *Chicken) Panicking() bool        { return c.panicTicks > 0 }
func (c *Chicken) Scale() float64 {
	if c.baby {
		return 0.5
	}
	return 1.0
}
func (c *Chicken) InteractText() string { return "Alimentar" }

func (c *Chicken) Heal(health float64, src world.HealingSource) { c.health += health }
func (c *Chicken) KnockBack(src mgl64.Vec3, f, h float64) {
	if c.self == nil { return }
	c.self.data.Vel = c.mc.KnockBack(src, f, h, c.self.data.Pos)
}
func (c *Chicken) Velocity() mgl64.Vec3       { return c.self.data.Vel }
func (c *Chicken) SetVelocity(v mgl64.Vec3)   { c.self.data.Vel = v }
func (c *Chicken) Speed() float64             { return 0.25 }
func (c *Chicken) SetSpeed(v float64)         {}
func (c *Chicken) AddEffect(e effect.Effect)  {}
func (c *Chicken) RemoveEffect(e effect.Type) {}
func (c *Chicken) Effects() []effect.Effect   { return nil }
func (c *Chicken) PistonImmovable() bool      { return false }
func (c *Chicken) PistonBreakable() bool      { return false }

func (c *Chicken) UUID() uuid.UUID {
	if c.self == nil { return uuid.UUID{} }
	return c.self.H().UUID()
}

func (c *Chicken) DeathPosition() (mgl64.Vec3, world.Dimension, bool) {
	if c.self == nil { return mgl64.Vec3{}, nil, c.Dead() }
	return c.self.Position(), c.self.tx.World().Dimension(), c.Dead()
}

func (c *Chicken) Variant() int32 {
	return int32(c.variant)
}

func (c *Chicken) UseOnEntity(tx *world.Tx, user item.User, held item.Stack) bool {
	if c.Dead() {
		return false
	}
	switch it := held.Item().(type) {
	case block.WheatSeeds, block.BeetrootSeeds, block.MelonSeeds, block.PumpkinSeeds:
		if c.baby {
			c.growthTicks += 2400 // 10%
			if c.growthTicks >= 24000 {
				c.grow(tx)
			}
			return true
		}
		if c.breedCooldown == 0 && c.loveTicks == 0 {
			c.loveTicks = 600
			for _, v := range tx.Viewers(c.self.Position()) {
				v.ViewEntityAction(c.self, InLoveAction{})
			}
			return true
		}
	case block.Flower:
		if it.Type == block.Dandelion() && c.baby {
			return true
		}
	}
	return false
}
