package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand"
)

// MovingBlock is a technical block used to represent a block that is currently being moved by a piston.
type MovingBlock struct {
	transparent
	empty

	// MovingBlock is the block that is being moved.
	MovingBlock world.Block
	// Facing is the direction the block is moving in.
	Facing cube.Face
	// Extending is true if the piston is extending.
	Extending bool
}

// EncodeNBT ...
func (m MovingBlock) EncodeNBT() map[string]any {
	name, properties := m.MovingBlock.EncodeBlock()
	extending := uint8(0)
	if m.Extending {
		extending = 1
	}
	return map[string]any{
		"id":         "MovingBlock",
		"blockState": map[string]any{"name": name, "states": properties},
		"extending":  extending,
	}
}

// DecodeNBT ...
func (m MovingBlock) DecodeNBT(data map[string]any) any {
	return m
}

// ScheduledTick restores the original block after the animation ends.
func (m MovingBlock) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	tx.SetBlock(pos, m.MovingBlock, nil)
}

// BreakInfo ...
func (m MovingBlock) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, nil)
}

// Model ...
func (m MovingBlock) Model() world.BlockModel {
	return model.Empty{}
}

// EncodeItem ...
func (m MovingBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:moving_block", 0
}

// EncodeBlock ...
func (m MovingBlock) EncodeBlock() (string, map[string]any) {
	return "minecraft:moving_block", nil
}

// MovingAction is a block action used to trigger the piston animation.
type MovingAction struct{}

func (MovingAction) BlockAction() {}
