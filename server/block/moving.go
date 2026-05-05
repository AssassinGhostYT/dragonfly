package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
)

// Moving is a technical block used to represent a block that is currently being moved by a piston.
type Moving struct {
	transparent
	empty

	// Piston is the position of the piston that moved this block.
	Piston cube.Pos
	// Moving is the block that is being moved.
	Moving world.Block
	// Expanding is true if the piston is extending.
	Expanding bool
}

// EncodeBlock ...
func (m Moving) EncodeBlock() (string, map[string]any) {
	return "minecraft:moving_block", nil
}

// Model ...
func (m Moving) Model() world.BlockModel {
	return model.Empty{}
}

// EncodeItem ...
func (m Moving) EncodeItem() (name string, meta int16) {
	return "minecraft:moving_block", 0
}

// EncodeNBT ...
func (m Moving) EncodeNBT() map[string]any {
	name, properties := m.Moving.EncodeBlock()
	extending := uint8(0)
	if m.Expanding {
		extending = 1
	}
	return map[string]any{
		"id":         "MovingBlock",
		"blockState": map[string]any{"name": name, "states": properties},
		"extending":  extending,
		"pistonPosX": int32(m.Piston.X()),
		"pistonPosY": int32(m.Piston.Y()),
		"pistonPosZ": int32(m.Piston.Z()),
	}
}

// DecodeNBT ...
func (m Moving) DecodeNBT(data map[string]any) any {
	return m
}

// allMoving ...
func allMoving() []world.Block {
	return []world.Block{Moving{Moving: Air{}}}
}
