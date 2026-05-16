package particle

import "image/color"

// HugeExplosion is a particle shown when TNT or a creeper explodes.
type HugeExplosion struct{ particle }

// EndermanTeleport is a particle that shows up when an enderman teleports.
type EndermanTeleport struct{ particle }

// SnowballPoof is a particle shown when a snowball collides with something.
type SnowballPoof struct{ particle }

// EggSmash is a particle shown when an egg smashes on something.
type EggSmash struct{ particle }

// Splash is a particle that shows up when a splash potion is splashed.
type Splash struct {
	particle

	// Colour is the colour that should be splashed.
	Colour color.RGBA
}

// Effect is a particle that shows up around an entity when it has effects on.
type Effect struct {
	particle

	// Colour is the colour of the particle.
	Colour color.RGBA
}

// EntityFlame is a particle shown when an entity is set on fire.
type EntityFlame struct{ particle }

// VibrationSignal is a particle shown when a sculk sensor detects a vibration.
type VibrationSignal struct {
	particle
	// Origin is the position where the vibration started.
	Origin mgl64.Vec3
	// Destination is the position where the vibration is traveling to.
	Destination mgl64.Vec3
}
