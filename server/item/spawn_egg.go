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

// UseOnBlock spawns the entity at the position of the block clicked.
func (s SpawnEgg) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool {
	if s.Entity == nil {
		return false
	}
	opts := world.EntitySpawnOpts{Position: pos.Side(face).Vec3Middle()}

	name := s.Entity.EncodeEntity()
	if name == "minecraft:silverfish" && NewSilverfish != nil {
		tx.AddEntity(NewSilverfish(opts))
		ctx.SubtractFromCount(1)
		return true
	}

	return false
}

// EncodeItem ...
func (s SpawnEgg) EncodeItem() (name string, meta int16) {
	// This is a bit tricky as different eggs have different names in Bedrock.
	// For now we only implement silverfish.
	return "minecraft:silverfish_spawn_egg", 0
}

// MaxCount ...
func (s SpawnEgg) MaxCount() int {
	return 64
}
