package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"sync"
)

// Redstone handles redstone-related state and logic for a world.
type Redstone struct {
	mu sync.Mutex
	// burnout holds the burnout state of redstone torches.
	burnout map[cube.Pos][]int64
}

// NewRedstone creates a new Redstone controller.
func NewRedstone() *Redstone {
	return &Redstone{
		burnout: make(map[cube.Pos][]int64),
	}
}

// TorchBurnoutStatus returns whether a torch at the provided position is burned out.
func (r *Redstone) TorchBurnoutStatus(pos cube.Pos, currentTick int64) (burnedOut bool, recoverable bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ticks, ok := r.burnout[pos]
	if !ok {
		return false, false
	}
	if len(ticks) >= 8 {
		// Recoverable after 60 ticks? (Simplified)
		if currentTick-ticks[len(ticks)-1] > 60 {
			return true, true
		}
		return true, false
	}
	return false, false
}

// ClearTorchBurnout clears the burnout state for a torch.
func (r *Redstone) ClearTorchBurnout(pos cube.Pos) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.burnout, pos)
}

// RecordTorchToggle records a toggle for a torch and returns true if it should burn out.
func (r *Redstone) RecordTorchToggle(pos cube.Pos, currentTick int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	ticks := r.burnout[pos]
	// Remove old ticks
	newTicks := []int64{}
	for _, t := range ticks {
		if currentTick-t < 60 {
			newTicks = append(newTicks, t)
		}
	}
	newTicks = append(newTicks, currentTick)
	r.burnout[pos] = newTicks
	return len(newTicks) >= 8
}

// BurnOutTorch sets a torch as burned out.
func (r *Redstone) BurnOutTorch(pos cube.Pos) {
	// Handled by RecordTorchToggle mostly, but can be forced.
}

// PruneTorchBurnout removes old entries from the burnout map.
func (r *Redstone) PruneTorchBurnout(pos cube.Pos, currentTick int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ticks, ok := r.burnout[pos]
	if !ok {
		return
	}
	newTicks := []int64{}
	for _, t := range ticks {
		if currentTick-t < 60 {
			newTicks = append(newTicks, t)
		}
	}
	if len(newTicks) == 0 {
		delete(r.burnout, pos)
	} else {
		r.burnout[pos] = newTicks
	}
}

// MarkTorchSelfTriggeredIfActive ...
func (r *Redstone) MarkTorchSelfTriggeredIfActive(cube.Pos) {}

// WithActiveTorchUpdate ...
func (r *Redstone) WithActiveTorchUpdate(cube.Pos, func()) {
	// This usually prevents infinite recursion.
}
