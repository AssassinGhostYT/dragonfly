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
	"math"
	"math/rand"
	"time"
)

// SulfurCube is a passive mob that can swallow blocks and change its physics.
type SulfurCube struct {
	brain     *mobsx.Brain
	navigator *mobsx.Navigator
	mc        *MovementComputer
	self      *Ent

	health float64
	size   int // 1: Small, 2: Large

	swallowedBlock world.Block
	pickupTimer    int
	fuse           int // Ticks until TNT explosion, -1 if not ignited
	jumpTicks      int
	growthTicks    int

	fromBucket bool
}

// NewSulfurCube creates a new Large Sulfur Cube.
func NewSulfurCube(opts world.EntitySpawnOpts) *world.EntityHandle {
	return opts.New(SulfurCubeType, &SulfurCube{health: 9, size: 2, fuse: -1})
}

// NewSulfurCubeSmall creates a new Small Sulfur Cube.
func NewSulfurCubeSmall(opts world.EntitySpawnOpts) *world.EntityHandle {
	return opts.New(SulfurCubeType, &SulfurCube{health: 4, size: 1, fuse: -1})
}

func (s *SulfurCube) Apply(data *world.EntityData) {
	s.mc = &MovementComputer{Gravity: 0.08, Drag: 0.02, StepHeight: 1.0}
	s.updatePhysics()
	data.Data = s
}

func (s *SulfurCube) updatePhysics() {
	if s.swallowedBlock == nil {
		s.mc.Gravity = 0.08
		s.mc.Drag = 0.02
		return
	}

	arch := s.Archetype()
	// Default base
	s.mc.Gravity = 0.08
	s.mc.Drag = 0.02

	switch arch {
	case "bouncy":
		// -2.0 Knockback Res, 0.9 Elasticity, 0.3 Friction, 0.01 Drag, Floats: Yes
		s.mc.Drag = 0.01
	case "explosive":
		// -1.0 Knockback Res, 0.5 Elasticity, 0.3 Friction, 0.3 Drag, Floats: Yes
		s.mc.Drag = 0.3
	case "fast_flat":
		// -1.0 Knockback Res, 0.5 Elasticity, 0.2 Friction, 0.01 Drag, Floats: No
		s.mc.Drag = 0.01
	case "fast_sliding":
		// 0.5 Knockback Res, 0.1 Elasticity, 0.05 Friction, 0.01 Drag, Floats: No
		s.mc.Drag = 0.01
	case "high_resistance":
		// 0.7 Knockback Res, 0.2 Elasticity, 1.0 Friction, 0.01 Drag, Floats: No
		s.mc.Drag = 0.01
	case "hot":
		// -1.0 Knockback Res, 0.5 Elasticity, 0.3 Friction, 0.1 Drag, Floats: Yes
		s.mc.Drag = 0.1
	case "light":
		// -1.0 Knockback Res, 1.0 Elasticity, 0.3 Friction, 1.8 Drag, Floats: Yes (Lower Gravity)
		s.mc.Gravity = 0.02
		s.mc.Drag = 1.8
	case "slow_bouncy":
		// 0.4 Knockback Res, 0.6 Elasticity, 0.3 Friction, 0.05 Drag, Floats: Yes
		s.mc.Drag = 0.05
	case "slow_flat":
		// 0.5 Knockback Res, 0.4 Elasticity, 0.4 Friction, 0.1 Drag, Floats: No (Higher Gravity)
		s.mc.Gravity = 0.12
		s.mc.Drag = 0.1
	case "slow_sliding":
		// 0.8 Knockback Res, 0.1 Elasticity, 0.05 Friction, 0.01 Drag, Floats: No
		s.mc.Drag = 0.01
	case "sticky":
		// -2.0 Knockback Res, 0.0 Elasticity, 2.0 Friction, 0.01 Drag, Floats: No
		s.mc.Drag = 0.01
	case "regular":
		// -1.0 Knockback Res, 0.5 Elasticity, 0.3 Friction, 0.1 Drag, Floats: Yes
		s.mc.Drag = 0.1
	}
}

func (s *SulfurCube) Archetype() string {
	if s.swallowedBlock == nil {
		return "none"
	}
	switch b := s.swallowedBlock.(type) {
	case block.Planks, block.Log:
		return "bouncy"
	case block.TNT:
		return "explosive"
	case block.Coral, block.Melon, block.Pumpkin, block.Froglight, block.HayBale:
		return "fast_flat"
	case block.BlueIce, block.PackedIce, block.Snow:
		return "fast_sliding"
	case block.SoulSand, block.SoulSoil:
		return "high_resistance"
	case block.Magma:
		return "hot"
	case block.Wool:
		return "light"
	case block.Stone, block.Cobblestone, block.Bricks, block.Sandstone, block.Deepslate:
		return "slow_bouncy"
	case block.Iron, block.Gold, block.Netherite, block.Copper, block.RawIron, block.RawGold, block.RawCopper, block.CoalOre, block.IronOre, block.GoldOre, block.CopperOre, block.DiamondOre:
		return "slow_flat" // Called "Medicine Ball" (Heavy)
	case block.Mushroom, block.Mud, block.Resin: // Approximations for Sliding
		return "slow_sliding"
	case block.Honeycomb:
		return "sticky"
	case block.Dirt, block.Sand, block.Gravel, block.Bedrock, block.Clay, block.Bone:
		return "regular"
	}
	return "regular"
}

func (s *SulfurCube) Tick(e *Ent, tx *world.Tx) *Movement {
	s.self = e
	if s.Dead() {
		return nil
	}

	if s.brain == nil {
		s.brain = mobsx.NewBrain()
		wBridge := WorldBridge{E: e}
		s.navigator = mobsx.NewNavigator(EntityBridge{E: e, tx: tx}, wBridge)
		s.navigator.Finder.Height = s.size
		s.navigator.Speed = 0.15

		s.brain.AddBehavior(behavior.NewWander(s.navigator, 40))
	}

	// Growth logic for small cubes
	if s.size == 1 {
		s.growthTicks++
		if s.growthTicks >= 24000 { // 20 mins
			s.grow(tx)
		}
	}

	// Pickup timer
	if s.pickupTimer > 0 {
		s.pickupTimer--
	}

	// TNT Archetype special logic
	if s.fuse >= 0 {
		s.fuse--
		if s.fuse%5 == 0 {
			tx.PlaySound(e.Position(), sound.TNT{})
		}
		if s.fuse <= 0 {
			s.explode(tx)
			return nil
		}
	}

	// Hot Archetype special logic
	if s.Archetype() == "hot" && tx.World().Time()%20 == 0 {
		for ent := range tx.EntitiesWithin(e.H().Type().BBox(e).Grow(0.5).Translate(e.Position())) {
			if ent.H().UUID() != e.H().UUID() {
				if l, ok := ent.(living); ok && !l.Dead() {
					if h, ok := ent.(interface{ Hurt(float64, world.DamageSource) (float64, bool) }); ok {
						h.Hurt(1.0, block.FireDamageSource{})
					}
				}
			}
		}
	}

	// Movement: Jumps (only if not immobilized by block)
	m := s.mc.TickMovement(e, e.data.Pos, e.data.Vel, e.data.Rot, tx)
	if s.swallowedBlock == nil {
		if m.OnGround {
			s.jumpTicks++
			if s.jumpTicks >= 20+rand.Intn(20) {
				e.data.Vel[1] = 0.42
				// Small forward boost if wandering
				if !s.navigator.Path.AtEnd() {
					yaw := float64(e.Rotation().Yaw())
					s.mc.Move(e, mgl64.Vec3{-math.Sin(yaw * math.Pi / 180), 0, math.Cos(yaw * math.Pi / 180)}.Mul(0.2))
				}
				s.jumpTicks = 0
				tx.PlaySound(e.Position(), sound.SlimeJump{})
			}
		}
	} else {
		// Immobilized but can be thrown/pushed
		s.navigator.Speed = 0
	}

	// Block absorption (Pickup items)
	if s.size == 2 && s.swallowedBlock == nil && s.pickupTimer <= 0 {
		for ent := range tx.EntitiesWithin(e.H().Type().BBox(e).Grow(1.0).Translate(e.Position())) {
			if itemEnt, ok := ent.(interface{ Behaviour() world.EntityBehaviour }); ok {
				if b, ok := itemEnt.Behaviour().(*ItemBehaviour); ok {
					if bl, ok := b.Item().Item().(world.Block); ok {
						// Simple check: most cubes are swallowable
						s.swallow(tx, bl)
						_ = ent.H().Close()
						break
					}
				}
			}
		}
	}

	wBridge := WorldBridge{E: e}
	s.navigator.Sync(wBridge)
	s.brain.Tick(EntityBridge{E: e, tx: tx}, wBridge)

	e.data.Pos, e.data.Vel = m.pos, m.vel
	return m
}

func (s *SulfurCube) swallow(tx *world.Tx, b world.Block) {
	if s.swallowedBlock != nil {
		s.eject(tx)
	}
	s.swallowedBlock = b
	s.updatePhysics()
	tx.PlaySound(s.self.Position(), sound.SlimeAttack{}) // Placeholder for absorb sound
	for _, v := range tx.Viewers(s.self.Position()) {
		v.ViewEntityState(s.self)
	}
}

func (s *SulfurCube) eject(tx *world.Tx) {
	if s.swallowedBlock == nil {
		return
	}
	// Drop block
	opts := world.EntitySpawnOpts{
		Position: s.self.Position().Add(mgl64.Vec3{0, 0.5, 0}),
		Velocity: mgl64.Vec3{rand.Float64()*0.2 - 0.1, 0.3, rand.Float64()*0.2 - 0.1},
	}
	tx.AddEntity(NewItem(opts, item.NewStack(s.swallowedBlock, 1)))
	s.swallowedBlock = nil
	s.pickupTimer = 100
	s.fuse = -1
	s.updatePhysics()
	tx.PlaySound(s.self.Position(), sound.ItemFrameRemoveItem{}) // Placeholder for eject sound
	for _, v := range tx.Viewers(s.self.Position()) {
		v.ViewEntityState(s.self)
	}
}

func (s *SulfurCube) grow(tx *world.Tx) {
	s.size = 2
	s.health = 9
	s.navigator.Finder.Height = 2
	for _, v := range tx.Viewers(s.self.Position()) {
		v.ViewEntityState(s.self)
	}
}

func (s *SulfurCube) explode(tx *world.Tx) {
	block.ExplosionConfig{Size: 3, ItemDropChance: 1.0}.Explode(tx, s.self.Position())
	_ = s.self.Close()
}

func (s *SulfurCube) Hurt(damage float64, src world.DamageSource) (n float64, v bool) {
	if s.Dead() {
		return 0, false
	}

	// Resistance if holding a block
	if s.swallowedBlock != nil {
		switch src.(type) {
		case AttackDamageSource, ProjectileDamageSource, block.ExplosionConfig:
			// "receive more recoil instead of damage"
			return 0, true 
		}
	}

	s.health -= damage
	if s.health > 0 && damage > 0 {
		s.self.tx.PlaySound(s.self.Position(), sound.SlimeHurt{})
		for _, v := range s.self.tx.Viewers(s.self.Position()) {
			v.ViewEntityAction(s.self, HurtAction{})
		}
	}

	if s.health <= 0 && s.self != nil {
		s.self.tx.AddParticle(s.self.Position(), particle.Evaporate{})
		s.self.tx.PlaySound(s.self.Position(), sound.SlimeDeath{})
		for _, v := range s.self.tx.Viewers(s.self.Position()) {
			v.ViewEntityAction(s.self, DeathAction{})
		}

		if s.size == 2 {
			// Divide into 2 small ones
			for i := 0; i < 2; i++ {
				opts := world.EntitySpawnOpts{
					Position: s.self.Position(),
					Velocity: mgl64.Vec3{rand.Float64()*0.2 - 0.1, 0.2, rand.Float64()*0.2 - 0.1},
				}
				s.self.tx.AddEntity(NewSulfurCubeSmall(opts))
			}
			// Drop swallowed block
			if s.swallowedBlock != nil {
				s.self.tx.AddEntity(NewItem(world.EntitySpawnOpts{Position: s.self.Position()}, item.NewStack(s.swallowedBlock, 1)))
			}
			// Drop XP
			for _, handle := range NewExperienceOrbs(s.self.Position(), rand.Intn(2)+1) {
				s.self.tx.AddEntity(handle)
			}
		}
		_ = s.self.Close()
	}
	return damage, true
}

func (s *SulfurCube) InteractText() string {
	if s.swallowedBlock != nil {
		return "Extraer Bloque"
	}
	return "Capturar en Cubo"
}

func (s *SulfurCube) Baby() bool { return s.size == 1 }
func (s *SulfurCube) Scale() float64 {
	if s.size == 1 {
		return 0.5
	}
	return 1.0
}

func (s *SulfurCube) Health() float64        { return s.health }
func (s *SulfurCube) MaxHealth() float64     { if s.size == 1 { return 4 }; return 9 }
func (s *SulfurCube) SetMaxHealth(v float64) { s.health = v }
func (s *SulfurCube) Dead() bool             { return s.health <= 0 }

func (s *SulfurCube) Heal(health float64, src world.HealingSource) { s.health += health }
func (s *SulfurCube) KnockBack(src mgl64.Vec3, f, h float64) {
	if s.self == nil { return }
	
	// Modifiers based on archetype
	horiz, vert := 1.0, 1.0
	if s.swallowedBlock != nil {
		switch s.Archetype() {
		case "bouncy": horiz, vert = 0.33, 0.07
		case "explosive": horiz, vert = 0.33, 0.06
		case "fast_flat": horiz, vert = 0.73, 0.06
		case "fast_sliding": horiz, vert = 0.53, 0.06
		case "high_resistance": horiz, vert = 0.33, 0.06
		case "hot": horiz, vert = 0.33, 0.06
		case "light": horiz, vert = 0.33, 0.12
		case "slow_bouncy": horiz, vert = 0.33, 0.16
		case "slow_flat": horiz, vert = 0.33, 0.07
		case "slow_sliding": horiz, vert = 0.33, 0.06
		case "sticky": horiz, vert = 0.33, 0.06
		case "regular": horiz, vert = 0.33, 0.06
		}
		// "Receive more recoil"
		f *= 1.5
		h *= 1.5
	}
	
	f *= horiz
	h *= vert
	s.self.data.Vel = s.mc.KnockBack(src, f, h, s.self.data.Pos)
}

func (s *SulfurCube) Velocity() mgl64.Vec3       { return s.self.data.Vel }
func (s *SulfurCube) SetVelocity(v mgl64.Vec3)   { s.self.data.Vel = v }
func (s *SulfurCube) Speed() float64             { if s.size == 1 { return 0.2 }; return 0.25 }
func (s *SulfurCube) SetSpeed(v float64)         {}
func (s *SulfurCube) AddEffect(e effect.Effect)  {}
func (s *SulfurCube) RemoveEffect(e effect.Type) {}
func (s *SulfurCube) Effects() []effect.Effect   { return nil }
func (s *SulfurCube) PistonImmovable() bool      { return false }
func (s *SulfurCube) PistonBreakable() bool      { return false }

func (s *SulfurCube) UUID() uuid.UUID {
	if s.self == nil { return uuid.UUID{} }
	return s.self.H().UUID()
}

func (s *SulfurCube) DeathPosition() (mgl64.Vec3, world.Dimension, bool) {
	if s.self == nil { return mgl64.Vec3{}, nil, s.Dead() }
	return s.self.Position(), s.self.tx.World().Dimension(), s.Dead()
}

// DisplayTileRuntimeID returns the block inside the Sulfur Cube.
func (s *SulfurCube) DisplayTileRuntimeID() (uint32, bool) {
	if s.swallowedBlock == nil {
		return 0, false
	}
	return world.BlockRuntimeID(s.swallowedBlock), true
}

func (s *SulfurCube) UseOnEntity(tx *world.Tx, user item.User, held item.Stack) bool {
	if s.Dead() {
		return false
	}
	switch i := held.Item().(type) {
	case item.Bucket:
		if i.Empty() && s.size == 2 {
			// Capture in bucket (simplified: just gives a message for now or logic to swap)
			return true
		}
	case item.Shears:
		if s.swallowedBlock != nil {
			s.eject(tx)
			return true
		}
	case item.FlintAndSteel, item.FireCharge:
		if s.Archetype() == "explosive" && s.fuse < 0 {
			s.fuse = 120 // 6 seconds
			tx.PlaySound(s.self.Position(), sound.TNT{})
			return true
		}
	case world.Block:
		if s.size == 2 {
			s.swallow(tx, i)
			return true
		}
	}
	return false
}

var SulfurCubeType sulfurCubeType

type sulfurCubeType struct{}

func (t sulfurCubeType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	e := &Ent{tx: tx, handle: handle, data: data}
	if s, ok := data.Data.(*SulfurCube); ok {
		s.self = e
	}
	return e
}
func (sulfurCubeType) EncodeEntity() string { return "minecraft:sulfur_cube" }

func (sulfurCubeType) BBox(e world.Entity) cube.BBox {
	if ent, ok := e.(*Ent); ok {
		if s, ok := ent.data.Data.(*SulfurCube); ok && s.size == 1 {
			return cube.Box(-0.245, 0, -0.245, 0.245, 0.49, 0.245)
		}
	}
	return cube.Box(-0.49, 0, -0.49, 0.49, 0.98, 0.49)
}
func (sulfurCubeType) DecodeNBT(m map[string]any, data *world.EntityData) {
	s := &SulfurCube{health: 9, size: 2, fuse: -1}
	if size, ok := m["Size"].(byte); ok {
		s.size = int(size)
		if s.size == 1 { s.health = 4 }
	}
	// Decode swallowed block here if needed
	s.Apply(data)
}
func (sulfurCubeType) EncodeNBT(data *world.EntityData) map[string]any {
	if s, ok := data.Data.(*SulfurCube); ok {
		return map[string]any{
			"Size": byte(s.size),
			"Fuse": int32(s.fuse),
		}
	}
	return nil
}
