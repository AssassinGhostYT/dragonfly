package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
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
		if strings.Contains(s.Name, "piston") {
			fmt.Printf("Name: %s, Properties: %v\n", s.Name, s.Properties)
		}
	}
}
