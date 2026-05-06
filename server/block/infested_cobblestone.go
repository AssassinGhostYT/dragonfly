package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// InfestedCobblestone is a block that hides a silverfish. It looks identical to cobblestone.
type InfestedCobblestone struct {
	solid
	bassDrum
}

// BreakInfo ...
func (i InfestedCobblestone) BreakInfo() BreakInfo {
	return newBreakInfo(1, pickaxeHarvestable, pickaxeEffective, silkTouchOnlyDrop(i)).withBlastResistance(0.75).withBreakHandler(func(pos cube.Pos, tx *world.Tx, u item.User) {
		if u != nil {
			opts := world.EntitySpawnOpts{Position: pos.Vec3Centre()}
			tx.AddEntity(tx.World().EntityRegistry().Config().Silverfish(opts))
		}
	})
}

// EncodeItem ...
func (i InfestedCobblestone) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_cobblestone", 0
}

// EncodeBlock ...
func (i InfestedCobblestone) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_cobblestone", nil
}
