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

// StickyPiston is a block capable of pushing and pulling blocks.
type StickyPiston struct {
	solid
	transparent

	// Facing is the direction the piston faces.
	Facing cube.Face
	// Extended is true if the piston is currently extended.
	Extended bool
}

// BreakInfo ...
func (p StickyPiston) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, alwaysHarvestable, pickaxeEffective, oneOf(p))
}

// UseOnBlock ...
func (p StickyPiston) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
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
func (p StickyPiston) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if p.isPowered(pos, tx) {
		if !p.Extended {
			p.extend(pos, tx)
		}
	} else {
		if p.Extended {
			p.retract(pos, tx)
		}
	}
}

// extend handles the extension of the piston, pushing up to 12 blocks.
func (p StickyPiston) extend(pos cube.Pos, tx *world.Tx) bool {
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
	tx.SetBlock(pushPos, PistonArmCollision{Facing: p.Facing, Sticky: true}, nil)
	
	p.Extended = true
	tx.SetBlock(pos, p, nil)
	tx.PlaySound(pos.Vec3(), sound.PistonExtend{})

	for _, v := range tx.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, MovingAction{})
	}
	return true
}

// retract handles the retraction of the piston.
func (p StickyPiston) retract(pos cube.Pos, tx *world.Tx) {
	armPos := pos.Side(p.Facing)
	tx.SetBlock(armPos, Air{}, nil)

	pullPos := armPos.Side(p.Facing)
	b := tx.Block(pullPos)
	if !p.isImmovable(b) {
		if _, ok := b.(Air); !ok {
			mb := MovingBlock{MovingBlock: b, Facing: p.Facing.Opposite(), Extending: false}
			tx.SetBlock(armPos, mb, nil)
			tx.ScheduleBlockUpdate(armPos, mb, time.Millisecond*100)
			tx.SetBlock(pullPos, Air{}, nil)
		}
	}

	p.Extended = false
	tx.SetBlock(pos, p, nil)
	tx.PlaySound(pos.Vec3(), sound.PistonRetract{})

	for _, v := range tx.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, MovingAction{})
	}
}

// isImmovable returns true if the block cannot be pushed.
func (p StickyPiston) isImmovable(b world.Block) bool {
	switch b.(type) {
	case Bedrock, Obsidian, InvisibleBedrock, Barrier, Water, Lava, PistonArmCollision:
		return true
	}
	return false
}

// isPowered checks if the piston is powered by an adjacent redstone source.
func (p StickyPiston) isPowered(pos cube.Pos, tx *world.Tx) bool {
	return receivedPower(pos, tx) > 0
}

// EncodeItem ...
func (p StickyPiston) EncodeItem() (name string, meta int16) {
	return "minecraft:sticky_piston", 0
}

// EncodeBlock ...
func (p StickyPiston) EncodeBlock() (string, map[string]any) {
	return "minecraft:sticky_piston", map[string]any{"facing_direction": pistonFace(p.Facing)}
}

// EncodeNBT ...
func (p StickyPiston) EncodeNBT() map[string]any {
	state := uint8(0)
	if p.Extended {
		state = 2
	}
	return map[string]any{
		"id":     "PistonArm",
		"state":  state,
		"sticky": uint8(1),
	}
}

// DecodeNBT ...
func (p StickyPiston) DecodeNBT(data map[string]any) any {
	p.Extended = nbtconv.Uint8(data, "state") == 2
	return p
}

// allStickyPistons ...
func allStickyPistons() (pistons []world.Block) {
	for _, face := range cube.Faces() {
		pistons = append(pistons, StickyPiston{Facing: face})
	}
	return
}
