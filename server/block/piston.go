package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// Piston is a block capable of pushing blocks when powered by redstone.
type Piston struct {
	solid
	transparent

	// Facing is the direction the piston faces.
	Facing cube.Face
	// Extended is true if the piston is currently extended.
	Extended bool
}

// BreakInfo ...
func (p Piston) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, alwaysHarvestable, pickaxeEffective, oneOf(p))
}

// UseOnBlock ...
func (p Piston) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, p)
	if !used {
		return
	}
	// Pistons face towards the player when placed.
	p.Facing = calculateFace(user, pos)

	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (p Piston) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	powered := receivedPower(pos, tx) > 0
	if powered {
		if !p.Extended {
			if p.extend(pos, tx) {
				p.Extended = true
				tx.SetBlock(pos, p, nil)
				tx.PlaySound(pos.Vec3(), sound.PistonExtend{})
			}
		}
	} else {
		if p.Extended {
			p.Extended = false
			tx.SetBlock(pos, p, nil)
			p.retract(pos, tx)
			tx.PlaySound(pos.Vec3(), sound.PistonRetract{})
		}
	}
}

// extend handles the extension of the piston, pushing up to 12 blocks.
func (p Piston) extend(pos cube.Pos, tx *world.Tx) bool {
	dir := p.Facing
	pushPos := pos.Side(dir)

	blocks := []cube.Pos{}
	curr := pushPos
	for len(blocks) < 12 {
		b := tx.Block(curr)
		if _, ok := b.(Air); ok {
			break
		}
		if p.isImmovable(b) {
			return false
		}
		blocks = append(blocks, curr)
		curr = curr.Side(dir)
	}

	if len(blocks) >= 12 {
		b := tx.Block(curr)
		if len(b.Model().BBox(curr, tx)) > 0 {
			return false
		}
	}

	// Move blocks in reverse order to avoid overwriting.
	for i := len(blocks) - 1; i >= 0; i-- {
		b := tx.Block(blocks[i])
		mb := MovingBlock{MovingBlock: b, Facing: dir, Extending: true}
		target := blocks[i].Side(dir)
		tx.SetBlock(target, mb, nil)
		tx.ScheduleBlockUpdate(target, mb, time.Millisecond*100)
	}
	tx.SetBlock(pushPos, PistonArmCollision{Facing: p.Facing, Sticky: false}, nil)

	for _, v := range tx.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, MovingAction{})
	}
	return true
}

// retract handles the retraction of the piston.
func (p Piston) retract(pos cube.Pos, tx *world.Tx) {
	armPos := pos.Side(p.Facing)
	if _, ok := tx.Block(armPos).(PistonArmCollision); ok {
		tx.SetBlock(armPos, Air{}, nil)
	}
	
	for _, v := range tx.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, MovingAction{})
	}
}

// isImmovable returns true if the block cannot be pushed.
func (p Piston) isImmovable(b world.Block) bool {
	switch b.(type) {
	case Bedrock, Obsidian, InvisibleBedrock, Barrier, Water, Lava, PistonArmCollision:
		return true
	}
	return false
}

// EncodeItem ...
func (p Piston) EncodeItem() (name string, meta int16) {
	return "minecraft:piston", 0
}

// EncodeBlock ...
func (p Piston) EncodeBlock() (string, map[string]any) {
	return "minecraft:piston", map[string]any{"facing_direction": pistonFace(p.Facing)}
}

// EncodeNBT ...
func (p Piston) EncodeNBT() map[string]any {
	state := uint8(0) // Retracted
	if p.Extended {
		state = 2 // Extended
	}
	return map[string]any{
		"id":    "PistonArm",
		"state": state,
	}
}

// DecodeNBT ...
func (p Piston) DecodeNBT(data map[string]any) any {
	p.Extended = nbtconv.Uint8(data, "state") == 2
	return p
}

// allPistons ...
func allPistons() (pistons []world.Block) {
	for _, face := range cube.Faces() {
		pistons = append(pistons, Piston{Facing: face})
	}
	return
}
