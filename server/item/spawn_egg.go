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
}

// NewSilverfish is a function that can be used to create a new silverfish entity. It is set by the entity package.
var NewSilverfish func(opts world.EntitySpawnOpts) *world.EntityHandle

// NewZombie is a function that can be used to create a new zombie entity. It is set by the entity package.
var NewZombie func(opts world.EntitySpawnOpts) *world.EntityHandle

// UseOnBlock spawns the entity at the position of the block clicked.
func (s SpawnEgg) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool {
	if s.Entity == nil {
		return false
	}
	opts := world.EntitySpawnOpts{Position: pos.Side(face).Vec3Middle()}

	name := s.Entity.EncodeEntity()
	switch name {
	case "minecraft:silverfish":
		if NewSilverfish != nil {
			tx.AddEntity(NewSilverfish(opts))
			ctx.SubtractFromCount(1)
			return true
		}
	case "minecraft:zombie":
		if NewZombie != nil {
			tx.AddEntity(NewZombie(opts))
			ctx.SubtractFromCount(1)
			return true
		}
	}

	return false
}

// EncodeItem ...
func (s SpawnEgg) EncodeItem() (name string, meta int16) {
	if s.Entity != nil {
		return s.Entity.EncodeEntity() + "_spawn_egg", 0
	}
	return "minecraft:spawn_egg", 0
}

// MaxCount ...
func (s SpawnEgg) MaxCount() int {
	return 64
}
