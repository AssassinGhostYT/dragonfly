package entity

import (
	"github.com/AssassinGhostYT/MobsX-MC/api"
	"github.com/AssassinGhostYT/MobsX-MC/mmath"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// worldBridge implements MobsX-MC api.World.
type worldBridge struct {
	tx *world.Tx
}

func (w worldBridge) Block(pos mmath.Pos) api.Block {
	b := w.tx.Block(cube.Pos{pos.X(), pos.Y(), pos.Z()})
	return blockBridge{b: b}
}

func (w worldBridge) Entities() []api.Entity {
	return nil
}

// blockBridge implements MobsX-MC api.Block.
type blockBridge struct {
	b world.Block
}

func (b blockBridge) Name() string {
	if n, ok := b.b.(interface{ EncodeBlock() (string, map[string]any) }); ok {
		name, _ := n.EncodeBlock()
		return name
	}
	return "minecraft:air"
}

func (b blockBridge) Solid() bool {
	return b.Name() != "minecraft:air" 
}

func (b blockBridge) Passable() bool {
	return b.Name() == "minecraft:air"
}

// EntityBridge implements MobsX-MC api.Entity.
type EntityBridge struct {
	E  *Ent
	Tx *world.Tx
}

func (e EntityBridge) Position() [3]float64 {
	pos := e.E.Position()
	return [3]float64{pos.X(), pos.Y(), pos.Z()}
}

func (e EntityBridge) SetPosition(pos [3]float64) {
	e.E.data.Pos = [3]float64{pos[0], pos[1], pos[2]}
}

func (e EntityBridge) Rotation() [2]float32 {
	rot := e.E.Rotation()
	return [2]float32{float32(rot.Yaw()), float32(rot.Pitch())}
}

func (e EntityBridge) SetRotation(yaw, pitch float32) {
	e.E.data.Rot = cube.Rotation{float64(yaw), float64(pitch)}
}

func (e EntityBridge) ID() int64 {
	return 0
}

func (e EntityBridge) HideInBlock(pos mmath.Pos) {}

func (e EntityBridge) AlertOthers(rangeX, rangeY, rangeZ int) {}
