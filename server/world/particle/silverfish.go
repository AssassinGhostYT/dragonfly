package particle

import "github.com/df-mc/dragonfly/server/world"

// Cloud is a particle shown when a mob spawns or an entity is infested.
type Cloud struct{ particle }

// Death is a particle shown when a mob dies (same as cloud poof).
type Death struct{ particle }
