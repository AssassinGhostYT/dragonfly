package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

//go:embed block_states.nbt
var blockStateData []byte

type blockState struct {
	Name       string         `nbt:"name"`
	Properties map[string]any `nbt:"states"`
	Version    int32          `nbt:"version"`
}

func main() {
	dec := nbt.NewDecoder(bytes.NewBuffer(blockStateData))
	for {
		var s blockState
		if err := dec.Decode(&s); err != nil {
			break
		}
		if s.Name == "minecraft:piston" || s.Name == "minecraft:sticky_piston" || s.Name == "minecraft:piston_arm_collision" {
			fmt.Printf("Name: %s, Properties: %+v\n", s.Name, s.Properties)
		}
	}
}
