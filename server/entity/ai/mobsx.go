package ai

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Sensor represents a system that gathers information from the world for an entity.
type Sensor interface {
	// Scan performs the detection logic and returns true if it found something relevant.
	Scan(e world.Entity, tx *world.Tx) bool
}

// Behavior represents a specific goal or action a mob wants to perform.
type Behavior interface {
	// Priority returns the importance of this behavior (higher = more important).
	Priority() int
	// CanRun returns true if the behavior requirements are met.
	CanRun(e world.Entity, tx *world.Tx) bool
	// Run executes the behavior logic.
	Run(e world.Entity, tx *world.Tx)
}

// MobsX represents the main AI controller for an entity.
type MobsX struct {
	sensors   []Sensor
	behaviors []Behavior
	
	activeBehavior Behavior
}

// NewMobsX creates a new MobsX-MC AI controller.
func NewMobsX() *MobsX {
	return &MobsX{
		sensors:   []Sensor{},
		behaviors: []Behavior{},
	}
}

// AddSensor adds a detection system to the AI.
func (m *MobsX) AddSensor(s Sensor) {
	m.sensors = append(m.sensors, s)
}

// AddBehavior adds a goal system to the AI.
func (m *MobsX) AddBehavior(b Behavior) {
	m.behaviors = append(m.behaviors, b)
}

// Tick executes one cycle of the AI.
func (m *MobsX) Tick(e world.Entity, tx *world.Tx) {
	// 1. Update all sensors
	for _, s := range m.sensors {
		s.Scan(e, tx)
	}

	// 2. Decide which behavior to run based on priority
	var bestBehavior Behavior
	for _, b := range m.behaviors {
		if b.CanRun(e, tx) {
			if bestBehavior == nil || b.Priority() > bestBehavior.Priority() {
				bestBehavior = b
			}
		}
	}

	// 3. Run the best behavior
	if bestBehavior != nil {
		m.activeBehavior = bestBehavior
		bestBehavior.Run(e, tx)
	}
}
