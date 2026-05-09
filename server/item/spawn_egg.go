package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SpawnEgg is an item that spawns an entity when used on a block.
type SpawnEgg struct {
	// Entity is the type of entity to spawn.
	Entity world.EntityType
	// Meta is the metadata value used for the egg.
	Meta int16
}

// NewSilverfish is a function that can be used to create a new silverfish entity. It is set by the entity package.
var NewSilverfish func(opts world.EntitySpawnOpts) *world.EntityHandle

// NewZombie is a function that can be used to create a new zombie entity. It is set by the entity package.
var NewZombie func(opts world.EntitySpawnOpts) *world.EntityHandle

// NewZombieBaby is a function that can be used to create a new baby zombie entity. It is set by the entity package.
var NewZombieBaby func(opts world.EntitySpawnOpts) *world.EntityHandle

// NewChicken is a function that can be used to create a new chicken entity. It is set by the entity package.
var NewChicken func(opts world.EntitySpawnOpts) *world.EntityHandle

// NewChickenBaby is a function that can be used to create a new baby chicken entity. It is set by the entity package.
var NewChickenBaby func(opts world.EntitySpawnOpts) *world.EntityHandle

// UseOnBlock spawns the entity at the position of the block clicked.
func (s SpawnEgg) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool {
	if s.Entity == nil {
		return false
	}
	opts := world.EntitySpawnOpts{Position: pos.Side(face).Vec3Middle()}

	// Use the entity's spawn function if available, otherwise generic spawn (if we had one)
	// For now, we still use the hardcoded ones as we don't have a generic world.NewEntity in this version
	name := s.Entity.EncodeEntity()
	switch name {
	case "minecraft:silverfish":
		if NewSilverfish != nil {
			tx.AddEntity(NewSilverfish(opts))
		} else {
			return false
		}
	case "minecraft:zombie":
		if NewZombie != nil {
			tx.AddEntity(NewZombie(opts))
		} else {
			return false
		}
	case "minecraft:chicken":
		if NewChicken != nil {
			tx.AddEntity(NewChicken(opts))
		} else {
			return false
		}
	default:
		return false
	}

	ctx.SubtractFromCount(1)
	return true
}

// UseOnEntity spawns a baby entity if used on an adult of the same type.
func (s SpawnEgg) UseOnEntity(e world.Entity, tx *world.Tx, user User, ctx *UseContext) bool {
	if s.Entity == nil {
		return false
	}

	name := s.Entity.EncodeEntity()
	targetName := e.H().Type().EncodeEntity()

	if targetName == name {
		switch name {
		case "minecraft:zombie":
			if NewZombieBaby != nil {
				tx.AddEntity(NewZombieBaby(world.EntitySpawnOpts{Position: e.Position()}))
				ctx.SubtractFromCount(1)
				return true
			}
		case "minecraft:chicken":
			if NewChickenBaby != nil {
				tx.AddEntity(NewChickenBaby(world.EntitySpawnOpts{Position: e.Position()}))
				ctx.SubtractFromCount(1)
				return true
			}
		}
	}
	return false
}

// EncodeItem ...
func (s SpawnEgg) EncodeItem() (name string, meta int16) {
	return "minecraft:spawn_egg", s.Meta
}

// MaxCount ...
func (s SpawnEgg) MaxCount() int {
	return 64
}
