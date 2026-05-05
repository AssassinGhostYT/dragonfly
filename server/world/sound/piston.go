package sound

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// PistonExtend is a sound played when a piston extends.
type PistonExtend struct{ sound }

// PistonRetract is a sound played when a piston retracts.
type PistonRetract struct{ sound }

// Play ...
func (PistonExtend) Play(*world.World, mgl64.Vec3) {}

// Play ...
func (PistonRetract) Play(*world.World, mgl64.Vec3) {}
