package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// SculkShrieker is a block that shrieks and can summon a warden.
type SculkShrieker struct {
	empty
	transparent
	sourceWaterDisplacer
	// CanSummon specifies if the sculk shrieker can summon a warden.
	CanSummon bool
	// Shrieking specifies if the sculk shrieker is currently shrieking.
	Shrieking bool
}

// EntityStepOn ...
func (s SculkShrieker) EntityStepOn(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if !s.Shrieking {
		s.Shriek(tx, pos)
	}
}

// Shriek triggers the shrieking behavior of the block.
func (s SculkShrieker) Shriek(tx *world.Tx, pos cube.Pos) {
	s.Shrieking = true
	tx.SetBlock(pos, s, nil)

	// Play sound and particles.
	tx.PlaySound(pos.Vec3Centre(), sound.SculkShriekerShriek{})
	tx.AddParticle(pos.Vec3Centre(), particle.SculkShriekerShriek{})

	// Darkness effect for nearby players (16 block radius).
	for _, e := range tx.EntitiesWithin(cube.Box(
		float64(pos.X()-16), float64(pos.Y()-16), float64(pos.Z()-16),
		float64(pos.X()+16), float64(pos.Y()+16), float64(pos.Z()+16),
	), nil) {
		if l, ok := e.(interface{ AddEffect(e effect.Effect) }); ok {
			l.AddEffect(effect.New(effect.Darkness, 1, 12*time.Second))
		}
	}

	// Schedule to stop shrieking after 4.5 seconds.
	tx.ScheduleBlockUpdate(pos, s, 90*time.Second/20)
}

// ScheduledTick ...
func (s SculkShrieker) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	s.Shrieking = false
	tx.SetBlock(pos, s, nil)
}

// UseOnBlock ...
func (s SculkShrieker) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, ok := firstReplaceable(tx, pos, face, s)
	if !ok {
		return false
	}

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (s SculkShrieker) BreakInfo() BreakInfo {
	return newBreakInfo(3, alwaysHarvestable, hoeEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(s, 1)}
		}
		return nil
	}).withXPDropRange(5, 5).withBlastResistance(3)
}

// EncodeItem ...
func (SculkShrieker) EncodeItem() (name string, meta int16) {
	return "minecraft:sculk_shrieker", 0
}

// EncodeBlock ...
func (s SculkShrieker) EncodeBlock() (string, map[string]any) {
	return "minecraft:sculk_shrieker", map[string]any{
		"can_summon": boolByte(s.CanSummon),
		"active":     boolByte(s.Shrieking),
	}
}

// allSculkShriekers ...
func allSculkShriekers() (b []world.Block) {
	for _, c := range []bool{true, false} {
		for _, s := range []bool{true, false} {
			b = append(b, SculkShrieker{CanSummon: c, Shrieking: s})
		}
	}
	return
}
