package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// InfestedStone is a block that hides a silverfish. It looks identical to stone.
type InfestedStone struct {
	solid
	bassDrum

	// Type is the type of block that is infested.
	Type InfestedType
}

// BreakInfo ...
func (i InfestedStone) BreakInfo() BreakInfo {
	return newBreakInfo(0.75, alwaysHarvestable, nothingEffective, nil).withBreakHandler(func(pos cube.Pos, tx *world.Tx, u item.User) {
		// Wiki: Spawns silverfish when broken.
		// We'll spawn it here if it's not creative.
		if g, ok := u.(interface{ GameMode() world.GameMode }); !ok || !g.GameMode().CreativeInventory() {
			// Logic to spawn Silverfish at pos.Vec3Centre()
		}
	})
}

// EncodeItem ...
func (i InfestedStone) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_stone", 0
}

// EncodeBlock ...
func (i InfestedStone) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_stone", map[string]any{"monster_egg_stone_type": i.Type.String()}
}

// InfestedType ...
type InfestedType struct {
	infestedType
}

type infestedType uint8

func (i infestedType) String() string {
	switch i {
	case 0: return "stone"
	case 1: return "cobblestone"
	case 2: return "stone_brick"
	case 3: return "mossy_stone_brick"
	case 4: return "cracked_stone_brick"
	case 5: return "chiseled_stone_brick"
	}
	return "stone"
}

func InfestedStoneBlock() InfestedType { return InfestedType{0} }
func InfestedCobblestone() InfestedType { return InfestedType{1} }
func InfestedStoneBrick() InfestedType { return InfestedType{2} }
func InfestedMossyStoneBrick() InfestedType { return InfestedType{3} }
func InfestedCrackedStoneBrick() InfestedType { return InfestedType{4} }
func InfestedChiseledStoneBrick() InfestedType { return InfestedType{5} }

func allInfestedTypes() []InfestedType {
	return []InfestedType{
		InfestedStoneBlock(),
		InfestedCobblestone(),
		InfestedStoneBrick(),
		InfestedMossyStoneBrick(),
		InfestedCrackedStoneBrick(),
		InfestedChiseledStoneBrick(),
	}
}

// allInfestedStones ...
func allInfestedStones() (stones []world.Block) {
	for _, t := range allInfestedTypes() {
		stones = append(stones, InfestedStone{Type: t})
	}
	return
}
