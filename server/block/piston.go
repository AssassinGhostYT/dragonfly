package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// Piston is a block capable of pushing blocks when powered by redstone.
type Piston struct {
	solid
	transparent

	// Facing is the direction the piston faces.
	Facing cube.Face
}

// BreakInfo ...
func (p Piston) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, pickaxeEffective, oneOf(p))
}

// UseOnBlock ...
func (p Piston) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, p)
	if !used {
		return
	}
	p.Facing = calculateFace(user, pos)

	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (p Piston) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if p.isPowered(pos, tx) {
		if !p.neighbouringArm(pos, tx) {
			p.extend(pos, tx)
		}
	} else {
		if p.neighbouringArm(pos, tx) {
			p.retract(pos, tx)
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
	tx.SetBlock(armPos, Air{}, nil)
	for _, v := range tx.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, MovingAction{})
	}
}

// neighbouringArm returns true if a piston arm is present in front of the piston.
func (p Piston) neighbouringArm(pos cube.Pos, tx *world.Tx) bool {
	side := pos.Side(p.Facing)
	if arm, ok := tx.Block(side).(PistonArmCollision); ok {
		return arm.Facing == p.Facing
	}
	return false
}

// isImmovable returns true if the block cannot be pushed.
func (p Piston) isImmovable(b world.Block) bool {
	switch b.(type) {
	case Bedrock, Obsidian, InvisibleBedrock, Barrier, Water, Lava:
		return true
	}
	return false
}

// isPowered checks if the piston is powered by an adjacent redstone source.
func (p Piston) isPowered(pos cube.Pos, tx *world.Tx) bool {
	return receivedPower(pos, tx) > 0
}

// EncodeItem ...
func (p Piston) EncodeItem() (name string, meta int16) {
	return "minecraft:piston", 0
}

// EncodeBlock ...
func (p Piston) EncodeBlock() (string, map[string]any) {
	return "minecraft:piston", map[string]any{"facing_direction": int32(p.Facing)}
}

// allPistons ...
func allPistons() (pistons []world.Block) {
	for _, face := range cube.Faces() {
		pistons = append(pistons, Piston{Facing: face})
	}
	return
}
