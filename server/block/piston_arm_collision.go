package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// PistonArmCollision is a block that is used when a piston is extended and colliding with a block.
type PistonArmCollision struct {
	empty
	transparent
	sourceWaterDisplacer

	// Facing represents the direction the piston is facing.
	Facing cube.Face
	// Sticky is true if the piston is sticky.
	Sticky bool
}

// PistonImmovable ...
func (PistonArmCollision) PistonImmovable() bool {
	return true
}

// SideClosed ...
func (PistonArmCollision) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// EncodeBlock ...
func (c PistonArmCollision) EncodeBlock() (string, map[string]any) {
	name := "minecraft:piston_arm_collision"
	if c.Sticky {
		name = "minecraft:sticky_piston_arm_collision"
	}
	return name, map[string]any{"facing_direction": pistonFace(c.Facing)}
}

// BreakInfo ...
func (c PistonArmCollision) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, alwaysHarvestable, pickaxeEffective, simpleDrops()).withBreakHandler(func(pos cube.Pos, tx *world.Tx, u item.User) {
		pistonPos := pos.Side(c.Facing.Opposite())
		if p, ok := tx.Block(pistonPos).(Piston); ok {
			tx.SetBlock(pistonPos, nil, nil)
			if g, ok := u.(interface {
				GameMode() world.GameMode
			}); !ok || !g.GameMode().CreativeInventory() {
				dropItem(tx, item.NewStack(Piston{Sticky: p.Sticky}, 1), pos.Vec3Centre())
			}
		}
	})
}

// NeighbourUpdateTick ...
func (c PistonArmCollision) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if _, ok := tx.Block(pos.Side(c.Facing.Opposite())).(Piston); !ok {
		tx.SetBlock(pos, nil, nil)
	}
}

// allPistonArms ...
func allPistonArms() (pistonArmCollisions []world.Block) {
	for _, f := range cube.Faces() {
		pistonArmCollisions = append(pistonArmCollisions, PistonArmCollision{Facing: f, Sticky: false})
		pistonArmCollisions = append(pistonArmCollisions, PistonArmCollision{Facing: f, Sticky: true})
	}
	return
}
