package entity

import (
	"github.com/AssassinGhostYT/MobsX-MC/api"
	"github.com/AssassinGhostYT/MobsX-MC/mmath"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
	"unsafe"
)

// WorldBridge implements MobsX-MC api.World.
type WorldBridge struct {
	E *Ent
}

func (w WorldBridge) Block(pos mmath.Pos) api.Block {
	b := w.E.tx.Block(cube.Pos{pos.X(), pos.Y(), pos.Z()})
	return blockBridge{b: b}
}

func (w WorldBridge) Entities() []api.Entity {
	var ents []api.Entity
	for e := range w.E.tx.Entities() {
		ents = append(ents, EntityBridge{E: e, tx: w.E.tx})
	}
	return ents
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
	E  world.Entity
	tx *world.Tx
}

func (e EntityBridge) Position() [3]float64 {
	pos := e.E.Position()
	return [3]float64{pos.X(), pos.Y(), pos.Z()}
}

func (e EntityBridge) SetPosition(pos [3]float64) {
	if ent, ok := e.E.(*Ent); ok {
		current := ent.data.Pos
		target := mgl64.Vec3{pos[0], pos[1], pos[2]}
		diff := target.Sub(current)

		// For small movements, use velocity to make it smooth.
		if diff.Len() < 1.0 {
			ent.data.Vel = mgl64.Vec3{diff.X() * 0.8, ent.data.Vel.Y(), diff.Z() * 0.8}
		} else {
			// For large movements, teleport.
			ent.data.Pos = target
		}

		// Vertical boost for jumping. Only if we actually need to go up and we are not already jumping high.
		if target.Y()-current.Y() > 0.5 && ent.data.Vel.Y() < 0.1 {
			ent.data.Vel = mgl64.Vec3{ent.data.Vel.X(), 0.42, ent.data.Vel.Z()}
		}
	}
}

func (e EntityBridge) Rotation() [2]float32 {
	rot := e.E.Rotation()
	return [2]float32{float32(rot.Yaw()), float32(rot.Pitch())}
}

func (e EntityBridge) SetRotation(yaw, pitch float32) {
	if ent, ok := e.E.(*Ent); ok {
		ent.data.Rot = cube.Rotation{float64(yaw), float64(pitch)}
	}
}

func (e EntityBridge) ID() int64 {
	return int64(uintptr(unsafe.Pointer(e.E.H())))
}

func (e EntityBridge) IsPlayer() bool {
	if p, ok := e.E.(interface {
		GameMode() world.GameMode
	}); ok {
		return p.GameMode().AllowsTakingDamage()
	}
	return false
}

func (e EntityBridge) HeldItem() (name string, meta int16) {
	if p, ok := e.E.(interface {
		HeldItems() (item.Stack, item.Stack)
	}); ok {
		held, _ := p.HeldItems()
		return held.Item().EncodeItem()
	}
	return "", 0
}

func (e EntityBridge) HideInBlock(pos mmath.Pos) {
	if ent, ok := e.E.(*Ent); ok {
		b := e.tx.Block(cube.Pos{pos.X(), pos.Y(), pos.Z()})
		var infested world.Block
		if n, ok := b.(interface {
			EncodeBlock() (string, map[string]any)
		}); ok {
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
			e.tx.SetBlock(cube.Pos{pos.X(), pos.Y(), pos.Z()}, infested, nil)
			e.tx.AddParticle(cube.Pos{pos.X(), pos.Y(), pos.Z()}.Vec3Centre(), particle.Evaporate{})
			ent.Close()
		}
	}
}

func (e EntityBridge) Attack(target api.Entity, damage float64) {
	if t, ok := target.(EntityBridge); ok {
		if mgl64.Vec3(e.Position()).Sub(mgl64.Vec3(t.Position())).Len() > 2.0 {
			return
		}
		if p, ok := t.E.(interface {
			Hurt(damage float64, src world.DamageSource) (n float64, v bool)
		}); ok {
			p.Hurt(damage, AttackDamageSource{Attacker: e.E})
			for _, v := range e.tx.Viewers(e.E.Position()) {
				v.ViewEntityAction(e.E, SwingArmAction{})
			}
		}
	}
}

func (e EntityBridge) AlertOthers(rangeX, rangeY, rangeZ int) {
	pos := e.E.Position()
	center := cube.Pos{int(pos.X()), int(pos.Y()), int(pos.Z())}
	spawned := 0

	// Alert nearby already spawned silverfish
	for ent := range e.tx.Entities() {
		if s, ok := ent.(*Ent); ok {
			if silver, ok := s.data.Data.(*Silverfish); ok {
				if silver.self != nil && silver.self != e.E && silver.self.Position().Sub(pos).Len() < 16 {
					if silver.alerted != nil {
						silver.alerted.Alerted = true
					}
				}
			}
		}
	}

	for x := -rangeX / 2; x <= rangeX/2; x++ {
		for y := -rangeY / 2; y <= rangeY/2; y++ {
			for z := -rangeZ / 2; z <= rangeZ/2; z++ {
				checkPos := center.Add(cube.Pos{x, y, z})
				b := e.tx.Block(checkPos)

				var normal world.Block
				if n, ok := b.(interface {
					EncodeBlock() (string, map[string]any)
				}); ok {
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
					e.tx.SetBlock(checkPos, block.Air{}, nil)
					e.tx.AddParticle(checkPos.Vec3Centre(), particle.Evaporate{})
					opts := world.EntitySpawnOpts{Position: checkPos.Vec3Centre()}
					e.tx.AddEntity(NewSilverfish(opts))

					spawned++
					if spawned >= 3 {
						return
					}
				}
			}
		}
	}
}
