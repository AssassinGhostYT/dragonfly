package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// redstoneSource represents a block that can provide redstone power.
type redstoneSource interface {
	// WeakPower returns the weak redstone power level emitted by the block in the given direction.
	WeakPower(pos cube.Pos, face cube.Face, tx *world.Tx) int
	// StrongPower returns the strong redstone power level emitted by the block in the given direction.
	StrongPower(pos cube.Pos, face cube.Face, tx *world.Tx) int
}

// receivedPower returns the highest redstone power level received at the position passed.
func receivedPower(pos cube.Pos, tx *world.Tx) int {
	max := 0
	for _, face := range cube.Faces() {
		side := pos.Side(face)
		b := tx.Block(side)
		if s, ok := b.(redstoneSource); ok {
			p := s.WeakPower(side, face.Opposite(), tx)
			if p > max {
				max = p
			}
		}
	}
	return max
}

// isEnergized checks if a block is receiving a "strong" signal that would power its neighbors.
// This is used for redstone torches and blocks that need to transmit power through themselves.
func isEnergized(pos cube.Pos, tx *world.Tx) bool {
	// For now, let's keep it simple: if it's receiving power from any side.
	// In the future, we'll distinguish between block power and wire power.
	return receivedPower(pos, tx) > 0
}
