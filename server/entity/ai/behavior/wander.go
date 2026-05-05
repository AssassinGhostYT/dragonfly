package behavior

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand/v2"
)

// WanderBehavior makes an entity walk around aimlessly.
type WanderBehavior struct {
	// Radius is the maximum distance from the start position the entity can wander.
	Radius int
}

// Priority returns the priority of the wander behavior.
func (WanderBehavior) Priority() int {
	return 1 // Low priority
}

// CanRun returns true if the entity is not doing anything more important.
func (WanderBehavior) CanRun(e world.Entity, tx *world.Tx) bool {
	return true
}

// Run executes the wander logic.
func (w WanderBehavior) Run(e world.Entity, tx *world.Tx) {
	// Logic to pick a random spot and move there.
	// This will be expanded once we have the controller.
}
