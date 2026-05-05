package entity

import (
	"github.com/AssassinGhostYT/MobsX-MC/api"
	"github.com/AssassinGhostYT/MobsX-MC/internal/math"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// worldBridge implements api.World for Dragonfly.
type worldBridge struct {
	tx *world.Tx
}

func (w worldBridge) Block(pos math.Pos) api.Block {
	return blockBridge{b: w.tx.Block(cube.Pos{pos.X(), pos.Y(), pos.Z()})}
}

func (w worldBridge) Entities() []api.Entity {
	entities := []api.Entity{}
	for _, h := range w.tx.Entities() {
		if e, ok := w.tx.Entity(h); ok {
			entities = append(entities, EntityBridge{E: e})
		}
	}
	return entities
}

// blockBridge implements api.Block for Dragonfly.
type blockBridge struct {
	b world.Block
}

func (b blockBridge) Solid() bool {
	if _, ok := b.b.Model().(interface{ Solid() bool }); ok {
		return true
	}
	return false
}

func (b blockBridge) Passable() bool {
	return !b.Solid()
}

// EntityBridge can be used to wrap a Dragonfly entity.
type EntityBridge struct {
	E world.Entity
}

func (b EntityBridge) Position() [3]float64 {
	p := b.E.Position()
	return [3]float64{p.X(), p.Y(), p.Z()}
}

func (b EntityBridge) SetPosition(pos [3]float64) {
	if e, ok := b.E.(*Ent); ok {
		current := e.Position()
		// Set velocity to reach the target in one tick.
		e.data.Vel = mgl64.Vec3{pos[0] - current.X(), pos[1] - current.Y(), pos[2] - current.Z()}
	}
}

func (b EntityBridge) Rotation() [2]float32 {
	r := b.E.Rotation()
	return [2]float32{float32(r.Yaw()), float32(r.Pitch())}
}

func (b EntityBridge) SetRotation(yaw, pitch float32) {
	if e, ok := b.E.(*Ent); ok {
		e.data.Rot = cube.Rotation{float64(yaw), float64(pitch)}
	}
}

func (b EntityBridge) ID() int64 {
	// A unique ID is needed for the sensor to ignore self.
	// Since handle is unique, we can use it.
	return 0 // TODO: use something real if needed
}
