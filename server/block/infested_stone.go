package block

// InfestedStone is a block that hides a silverfish. It looks identical to stone.
type InfestedStone struct {
	solid
	bassDrum
}

// BreakInfo ...
func (i InfestedStone) BreakInfo() BreakInfo {
	return newBreakInfo(0.75, pickaxeHarvestable, pickaxeEffective, silkTouchOnlyDrop(i)).withBlastResistance(0.75).withBreakHandler(func(pos cube.Pos, tx *world.Tx, u item.User) {
		if u == nil {
			return
		}
		if _, ok := u.HeldItems(); ok {
			// Silk touch is handled by Drops, so here we only spawn if silk touch wasn't used.
			// However, Dragonfly doesn't pass enchantments to BreakHandler.
			// For now, we assume if u is in survival and breaks it, we spawn.
			opts := world.EntitySpawnOpts{Position: pos.Vec3Centre()}
			tx.AddEntity(tx.World().EntityRegistry().Config().Silverfish(opts))
		}
	})
}

// EncodeItem ...
func (i InfestedStone) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_stone", 0
}

// EncodeBlock ...
func (i InfestedStone) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_stone", nil
}
