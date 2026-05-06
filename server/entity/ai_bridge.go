package entity

import (
	"github.com/AssassinGhostYT/MobsX-MC/api"
	"github.com/AssassinGhostYT/MobsX-MC/mmath"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// worldBridge implements MobsX-MC api.World.
type worldBridge struct {
	E *Ent
}

func (w worldBridge) Block(pos mmath.Pos) api.Block {
	// Obtenemos la transacción actual de la entidad. Dragonfly actualiza w.E.tx en cada tick.
	b := w.E.tx.Block(cube.Pos{pos.X(), pos.Y(), pos.Z()})
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
	if n, ok := b.b.(interface {
		EncodeBlock() (string, map[string]any)
	}); ok {
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
	E *Ent
}

func (e EntityBridge) Position() [3]float64 {
	pos := e.E.Position()
	return [3]float64{pos.X(), pos.Y(), pos.Z()}
}

func (e EntityBridge) SetPosition(pos [3]float64) {
	// En lugar de teletransportar, calculamos la velocidad necesaria para llegar ahí.
	// Esto permite que las físicas de Dragonfly y las colisiones funcionen.
	current := e.E.data.Pos
	dx, dy, dz := pos[0]-current.X(), pos[1]-current.Y(), pos[2]-current.Z()

	// Aplicamos una velocidad suave hacia el objetivo.
	e.E.data.Vel = mgl64.Vec3{dx, dy, dz}
}

func (e EntityBridge) Rotation() [2]float32 {
	rot := e.E.Rotation()
	return [2]float32{float32(rot.Yaw()), float32(rot.Pitch())}
}

func (e EntityBridge) SetRotation(yaw, pitch float32) {
	e.E.data.Rot = cube.Rotation{float64(yaw), float64(pitch)}
}

func (e EntityBridge) ID() int64 {
	return int64(e.E.H().UUID().ID())
}

func (e EntityBridge) HideInBlock(pos mmath.Pos) {
	b := e.E.tx.Block(cube.Pos{pos.X(), pos.Y(), pos.Z()})
	var infested world.Block
	if n, ok := b.(interface{ EncodeBlock() (string, map[string]any) }); ok {
		name, _ := n.EncodeBlock()
		switch name {
		case "minecraft:stone":
			infested = block.InfestedStone{}
		case "minecraft:cobblestone":
			infested = block.InfestedCobblestone{}
		case "minecraft:deepslate":
			infested = block.InfestedDeepslate{}
		case "minecraft:stone_bricks":
			infested = block.InfestedStoneBricks{Type: block.NormalStoneBricks()}
		case "minecraft:mossy_stone_bricks":
			infested = block.InfestedStoneBricks{Type: block.MossyStoneBricks()}
		case "minecraft:cracked_stone_bricks":
			infested = block.InfestedStoneBricks{Type: block.CrackedStoneBricks()}
		case "minecraft:chiseled_stone_bricks":
			infested = block.InfestedStoneBricks{Type: block.ChiseledStoneBricks()}
		}
	}

	if infested != nil {
		e.E.tx.SetBlock(cube.Pos{pos.X(), pos.Y(), pos.Z()}, infested, nil)
		e.E.Close()
	}
}

func (e EntityBridge) AlertOthers(rangeX, rangeY, rangeZ int) {
	pos := e.E.Position()
	center := cube.Pos{int(pos.X()), int(pos.Y()), int(pos.Z())}
	spawned := 0

	for x := -rangeX / 2; x <= rangeX/2; x++ {
		for y := -rangeY / 2; y <= rangeY/2; y++ {
			for z := -rangeZ / 2; z <= rangeZ/2; z++ {
				checkPos := center.Add(cube.Pos{x, y, z})
				b := e.E.tx.Block(checkPos)

				var normal world.Block
				if n, ok := b.(interface{ EncodeBlock() (string, map[string]any) }); ok {
					name, _ := n.EncodeBlock()
					switch name {
					case "minecraft:infested_stone":
						normal = block.Stone{}
					case "minecraft:infested_cobblestone":
						normal = block.Cobblestone{}
					case "minecraft:infested_deepslate":
						normal = block.Deepslate{}
					case "minecraft:infested_stone_bricks":
						normal = block.StoneBricks{Type: block.NormalStoneBricks()}
					case "minecraft:infested_mossy_stone_bricks":
						normal = block.StoneBricks{Type: block.MossyStoneBricks()}
					case "minecraft:infested_cracked_stone_bricks":
						normal = block.StoneBricks{Type: block.CrackedStoneBricks()}
					case "minecraft:infested_chiseled_stone_bricks":
						normal = block.StoneBricks{Type: block.ChiseledStoneBricks()}
					}
				}

				if normal != nil {
					// Convertimos el bloque infestado en uno normal y sacamos el Silverfish.
					e.E.tx.SetBlock(checkPos, normal, nil)
					opts := world.EntitySpawnOpts{Position: checkPos.Vec3Centre()}
					e.E.tx.AddEntity(NewSilverfish(opts))

					// Limitamos a un máximo de 3 Silverfish por cada grito de ayuda para evitar lag masivo.
					spawned++
					if spawned >= 3 {
						return
					}
				}
			}
		}
	}
}
