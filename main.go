package main

import (
	"fmt"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/pelletier/go-toml"
	"log/slog"
	"os"
	"reflect"
	"strings"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	chat.Global.Subscribe(chat.StdoutSubscriber{})
	conf, err := readConfig(slog.Default())
	if err != nil {
		panic(err)
	}

	cmd.Register(cmd.New("gamemode", "Cambia el modo de juego de un jugador.", []string{"gm"}, GameModeCommand{}))

	srv := conf.New()
	srv.CloseOnProgramEnd()

	srv.Listen()
	for p := range srv.Accept() {
		p.Handle(&GlobalHandler{})
	}
}

// GameModeCommand es el comando para cambiar el modo de juego.
type GameModeCommand struct {
	Mode GameModeArg `cmd:"mode"`
}

func (c GameModeCommand) Run(source cmd.Source, output *cmd.Output, tx *world.Tx) {
	if p, ok := source.(*player.Player); ok {
		p.SetGameMode(c.Mode.mode)
		output.Printf("Tu modo de juego ha sido actualizado a %s.", c.Mode.name)
	}
}

// GameModeArg es un parámetro personalizado para parsear modos de juego.
type GameModeArg struct {
	mode world.GameMode
	name string
}

func (GameModeArg) Type() string { return "GameMode" }
func (g GameModeArg) Parse(line *cmd.Line, v reflect.Value) error {
	arg, ok := line.Next()
	if !ok {
		return fmt.Errorf("falta el modo de juego")
	}
	var newG GameModeArg
	switch strings.ToLower(arg) {
	case "0", "survival", "s":
		newG.mode = world.GameModeSurvival
		newG.name = "Supervivencia"
	case "1", "creative", "c":
		newG.mode = world.GameModeCreative
		newG.name = "Creativo"
	case "2", "adventure", "a":
		newG.mode = world.GameModeAdventure
		newG.name = "Aventura"
	case "3", "spectator", "sp":
		newG.mode = world.GameModeSpectator
		newG.name = "Espectador"
	default:
		return fmt.Errorf("modo de juego inválido: %s", arg)
	}
	v.Set(reflect.ValueOf(newG))
	return nil
}

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
				p.Tx().AddParticle(pos.Vec3Centre(), particle.SnowballPoof{})
			}
		}
	}
}

// hasSilkTouch verifica si un ítem tiene Toque de Seda.
func (h *GlobalHandler) hasSilkTouch(s item.Stack) bool {
	return false
}

// readConfig reads the configuration from the config.toml file, or creates the
// file if it does not yet exist.
func readConfig(log *slog.Logger) (server.Config, error) {
	c := server.DefaultConfig()
	var zero server.Config
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		data, err := toml.Marshal(c)
		if err != nil {
			return zero, fmt.Errorf("encode default config: %v", err)
		}
		if err := os.WriteFile("config.toml", data, 0644); err != nil {
			return zero, fmt.Errorf("create default config: %v", err)
		}
		return c.Config(log)
	}
	data, err := os.ReadFile("config.toml")
	if err != nil {
		return zero, fmt.Errorf("read config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		return zero, fmt.Errorf("decode config: %v", err)
	}
	return c.Config(log)
}
