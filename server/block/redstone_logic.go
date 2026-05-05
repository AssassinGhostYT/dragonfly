package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// updateStrongRedstone updates all neighbours of the provided position.
func updateStrongRedstone(pos cube.Pos, tx *world.Tx) {
	for _, face := range cube.Faces() {
		updateRedstone(pos.Side(face), tx)
	}
}

// updateRedstone updates the block at the provided position if it's a redstone receiver.
func updateRedstone(pos cube.Pos, tx *world.Tx) {
	if r, ok := tx.Block(pos).(interface {
		RedstoneUpdate(pos cube.Pos, tx *world.Tx)
	}); ok {
		r.RedstoneUpdate(pos, tx)
	}
}

// updateDirectionalRedstone updates redstone receivers in a specific direction.
func updateDirectionalRedstone(pos cube.Pos, tx *world.Tx, face cube.Face) {
	sidePos := pos.Side(face)
	updateRedstone(sidePos, tx)
	// Strong power propagation through solid blocks
	if _, ok := tx.Block(sidePos).Model().(interface{ Solid() bool }); ok {
		for _, f := range cube.Faces() {
			if f == face.Opposite() {
				continue
			}
			updateRedstone(sidePos.Side(f), tx)
		}
	}
}

// redstoneUpdateCancelled returns whether redstone updates are currently cancelled for the provided position.
func redstoneUpdateCancelled(cube.Pos, *world.Tx) bool {
	return false
}

// maxRedstoneWirePower returns the highest power between the current and the block's power.
func maxRedstoneWirePower(b world.Block, power int) int {
	if wire, ok := b.(RedstoneWire); ok {
		if wire.Power > power {
			return wire.Power
		}
	}
	return power
}

// receivedRedstonePower returns the highest power level received by a block at the provided position.
func receivedRedstonePower(pos cube.Pos, tx *world.Tx, face cube.Face) bool {
	return tx.RedstonePower(pos, face, true) > 0
}

// calculateAnySidedFace calculates a face for blocks that can face any direction.
func calculateAnySidedFace(user interface{ Rotation() cube.Rotation }, pos cube.Pos, horizontal bool) cube.Face {
	if horizontal {
		return user.Rotation().Direction().Face().Opposite()
	}
	// For vertical placement
	// This is a simplified version.
	return user.Rotation().Direction().Face().Opposite()
}

// RedstoneUpdater ...
type RedstoneUpdater interface {
	RedstoneUpdate(pos cube.Pos, tx *world.Tx)
}

// RedstoneWireStepDowner ...
type RedstoneWireStepDowner interface {
	CanRedstoneWireStepDown(pos, from cube.Pos, tx *world.Tx) bool
}
