package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SpawnEgg is an item used to spawn entities.
type SpawnEgg struct {
	// Entity is the type of entity to spawn.
	EntityType world.EntityType
}

// MaxCount ...
func (s SpawnEgg) MaxCount() int {
	return 64
}

// UseOnBlock ...
func (s SpawnEgg) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool {
	// Simple spawn logic for demo
	return false
}

// EncodeItem ...
func (s SpawnEgg) EncodeItem() (name string, meta int16) {
	return "minecraft:spawn_egg", 31
}
