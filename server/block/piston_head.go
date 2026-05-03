package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
)

// PistonHead is the technical block that represents the head of an extended piston.
type PistonHead struct {
	transparent
	// Facing is the direction the head faces.
	Facing cube.Face
	// Sticky is true if the head belongs to a sticky piston.
	Sticky bool
}

// BreakInfo ...
func (p PistonHead) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, pickaxeEffective, nil)
}

// EncodeItem ...
func (p PistonHead) EncodeItem() (name string, meta int16) {
	return "minecraft:piston_head", 0
}

// EncodeBlock ...
func (p PistonHead) EncodeBlock() (string, map[string]any) {
	return "minecraft:piston_head", map[string]any{
		"facing_direction": int32(p.Facing),
		"piston_type_bit":  p.Sticky,
	}
}

// Model ...
func (p PistonHead) Model() world.BlockModel {
	// TODO: Implement proper model with bounding boxes for the head and arm.
	return model.Empty{}
}

// allPistonHeads ...
func allPistonHeads() (heads []world.Block) {
	for _, face := range cube.Faces() {
		heads = append(heads, PistonHead{Facing: face, Sticky: false})
		heads = append(heads, PistonHead{Facing: face, Sticky: true})
	}
	return
}
