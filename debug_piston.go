package main

import (
	"bytes"
	"fmt"
	"os"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

type blockState struct {
	Name       string         `nbt:"name"`
	Properties map[string]any `nbt:"states"`
	Version    int32          `nbt:"version"`
}

func main() {
	data, err := os.ReadFile("server/world/block_states.nbt")
	if err != nil {
		panic(err)
	}

	dec := nbt.NewDecoder(bytes.NewBuffer(data))
	for {
		var s blockState
		if err := dec.Decode(&s); err != nil {
			break
		}
		if s.Name == "minecraft:piston" || s.Name == "minecraft:sticky_piston" {
			fmt.Printf("Block: %s, Props: %v\n", s.Name, s.Properties)
		}
	}
}
