package main

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"strings"
)

// GlobalHandler maneja eventos globales para implementar lógica sin tocar los bloques.
type GlobalHandler struct {
	player.NopHandler
}

// HandleBlockBreak detecta cuando se rompe un bloque infestado para spawnear el Silverfish.
func (h *GlobalHandler) HandleBlockBreak(ctx *player.Context, pos cube.Pos, drops *[]item.Stack, xp *int) {
	p := ctx.Val()
	b := p.Tx().Block(pos)
	
	// Verificamos si el bloque es infestado por su nombre codificado.
	if n, ok := b.(interface{ EncodeBlock() (string, map[string]any) }); ok {
		name, _ := n.EncodeBlock()
		if strings.HasPrefix(name, "minecraft:infested_") {
			// Si no tiene Silk Touch (Toque de Seda), spawneamos el Silverfish.
			held, _ := p.HeldItems()
			if !h.hasSilkTouch(held) {
				opts := world.EntitySpawnOpts{Position: pos.Vec3Centre()}
				p.Tx().AddEntity(entity.NewSilverfish(opts))
			}
		}
	}
}

// hasSilkTouch verifica si un ítem tiene Toque de Seda.
func (h *GlobalHandler) hasSilkTouch(s item.Stack) bool {
	// Implementación simplificada para el ejemplo.
	// En producción se usaría item.Enchantments() y se buscaría el tipo correcto.
	return false 
}
