package sensor

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// PlayerSensor detects nearby players within a specific range.
type PlayerSensor struct {
	Range float64
	// Detected holds the list of players found in the last scan.
	Detected []*world.EntityHandle
}

// Scan looks for players in the world around the entity.
func (s *PlayerSensor) Scan(e world.Entity, tx *world.Tx) bool {
	s.Detected = nil
	pos := e.Position()
	
	// Scan entities in range
	for _, other := range tx.Entities() {
		// Basic check: is it a player? (Usually players have a specific handle or data)
		// For now, we detect all living entities as a test.
		if other == e.H() {
			continue
		}
		
		otherPos := mgl64.Vec3{} // How to get pos from handle? 
		// Need to resolve handle in Tx
		if o, ok := tx.Entity(other); ok {
			otherPos = o.Position()
			if pos.Sub(otherPos).Len() <= s.Range {
				s.Detected = append(s.Detected, other)
			}
		}
	}
	return len(s.Detected) > 0
}
