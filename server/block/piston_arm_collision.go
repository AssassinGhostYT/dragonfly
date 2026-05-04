package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
)

// PistonArmCollision is the collision block for the piston arm.
type PistonArmCollision struct {
	transparent
	// Facing is the direction the arm faces.
	Facing cube.Face
	// Sticky is true if the piston is sticky.
	Sticky bool
}

// BreakInfo ...
func (p PistonArmCollision) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, pickaxeEffective, nil)
}

// EncodeItem ...
func (p PistonArmCollision) EncodeItem() (name string, meta int16) {
	if p.Sticky {
		return "minecraft:sticky_piston_arm_collision", 0
	}
	return "minecraft:piston_arm_collision", 0
}

// EncodeBlock ...
func (p PistonArmCollision) EncodeBlock() (string, map[string]any) {
	if p.Sticky {
		return "minecraft:sticky_piston_arm_collision", map[string]any{"facing_direction": int32(p.Facing)}
	}
	return "minecraft:piston_arm_collision", map[string]any{"facing_direction": int32(p.Facing)}
}

// Model ...
func (p PistonArmCollision) Model() world.BlockModel {
	return model.Empty{}
}

// allPistonArms ...
func allPistonArms() (arms []world.Block) {
	for _, face := range cube.Faces() {
		arms = append(arms, PistonArmCollision{Facing: face, Sticky: false})
		arms = append(arms, PistonArmCollision{Facing: face, Sticky: true})
	}
	return
}
