package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"math/rand"
	"sync"
	"time"
)

var (
	// warningLevels tracks the warning level of each player.
	warningLevels = make(map[uuid.UUID]int)
	// cooldowns tracks the last time a player activated a shrieker.
	cooldowns = make(map[uuid.UUID]time.Time)
	// mu protects warningLevels and cooldowns.
	mu sync.Mutex
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

// Model ...
func (s SculkShrieker) Model() world.BlockModel {
	return model.SculkShrieker{}
}

// EntityStepOn ...
func (s SculkShrieker) EntityStepOn(pos cube.Pos, tx *world.Tx, e world.Entity) {
	// Wiki: "Un chillidor de sculk se activa cuando cualquier jugador se para sobre la parte negra en el centro del bloque"
	if _, ok := e.(interface{ GameMode() world.GameMode }); ok {
		s.activate(tx, pos, e)
	}
}

// activate attempts to activate the shrieker.
func (s SculkShrieker) activate(tx *world.Tx, pos cube.Pos, e world.Entity) {
	if s.Shrieking {
		return
	}

	uid := uuid.Nil
	if h, ok := e.(interface{ H() *world.EntityHandle }); ok {
		uid = h.H().UUID()
	}

	mu.Lock()
	level := warningLevels[uid]
	if s.CanSummon && level < 4 {
		level++
		warningLevels[uid] = level
	}
	mu.Unlock()

	s.shriek(tx, pos, level)
}

// shriek triggers the shrieking behavior of the block.
func (s SculkShrieker) shriek(tx *world.Tx, pos cube.Pos, warningLevel int) {
	s.Shrieking = true
	tx.SetBlock(pos, s, nil)

	// Wiki: "Los chillones de escoba empapados no emiten sonido al chillar."
	liquid, _ := tx.Liquid(pos)
	_, water := liquid.(Water)
	if !water {
		tx.PlaySound(pos.Vec3Centre(), sound.SculkShriekerShriek{})
	}
	tx.AddParticle(pos.Vec3Centre(), particle.SculkShriekerShriek{})

	if s.CanSummon {
		// Play warning sounds based on level
		switch warningLevel {
		case 1:
			tx.PlaySound(pos.Vec3Centre(), sound.WardenNearbyClose{})
		case 2:
			tx.PlaySound(pos.Vec3Centre(), sound.WardenNearbyCloser{})
		case 3:
			tx.PlaySound(pos.Vec3Centre(), sound.WardenNearbyClosest{})
		case 4:
			// Warden spawning logic would go here. For now, just the sound if it "fails".
			tx.PlaySound(pos.Vec3Centre(), sound.WardenSlightlyAngry{})
		}
	}

	// Schedule to stop shrieking and apply darkness after 90 ticks (4.5 seconds).
	tx.ScheduleBlockUpdate(pos, s, 90*time.Second/20)
}

// ScheduledTick ...
func (s SculkShrieker) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	s.Shrieking = false
	tx.SetBlock(pos, s, nil)

	// Wiki: "Después de que terminen los chillidos, todos los jugadores en modo Supervivencia o Aventura dentro de 40 bloques reciben el efecto Oscuridad durante 12 segundos."
	if s.CanSummon {
		for e := range tx.EntitiesWithin(cube.Box(
			float64(pos.X()-40), float64(pos.Y()-40), float64(pos.Z()-40),
			float64(pos.X()+40), float64(pos.Y()+40), float64(pos.Z()+40),
		)) {
			if l, ok := e.(interface {
				AddEffect(e effect.Effect)
				GameMode() world.GameMode
			}); ok {
				gm := l.GameMode()
				if !gm.CreativeInventory() && gm.Visible() && gm.HasCollision() {
					l.AddEffect(effect.New(effect.Darkness, 1, 12*time.Second))
				}
			}
		}
	}
}

// UseOnBlock ...
func (s SculkShrieker) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, ok := firstReplaceable(tx, pos, face, s)
	if !ok {
		return false
	}
	// Wiki: "los chillones colocados por los jugadores o los catalizadores sculk son completamente inertes e inofensivos."
	s.CanSummon = false

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (s SculkShrieker) BreakInfo() BreakInfo {
	return newBreakInfo(3, alwaysHarvestable, hoeEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		// Wiki: "pierde la capacidad de invocar guardianes, incluso si se extrae con Toque de Seda."
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(SculkShrieker{CanSummon: false}, 1)}
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
