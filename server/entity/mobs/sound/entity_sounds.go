package mobsound

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// sound implements the world.Sound interface.
type sound struct{}

// Play ...
func (sound) Play(*world.World, mgl64.Vec3) {}

// --- SILVERFISH ---

// SilverfishAmbient is a sound played randomly by a silverfish.
type SilverfishAmbient struct{ sound }

// SilverfishHurt is a sound played when a silverfish is hurt.
type SilverfishHurt struct{ sound }

// SilverfishDeath is a sound played when a silverfish dies.
type SilverfishDeath struct{ sound }

// SilverfishStep is a sound played when a silverfish walks.
type SilverfishStep struct{ sound }

// --- ZOMBIE ---

// ZombieAmbient is a sound played randomly by a zombie.
type ZombieAmbient struct{ sound }

// ZombieHurt is a sound played when a zombie is hurt.
type ZombieHurt struct{ sound }

// ZombieDeath is a sound played when a zombie dies.
type ZombieDeath struct{ sound }

// ZombieStep is a sound played when a zombie steps on a block.
type ZombieStep struct{ sound }
