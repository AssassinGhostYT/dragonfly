package item

// Redstone is an item that can be placed as redstone dust.
type Redstone struct{}

// EncodeItem ...
func (Redstone) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone", 0
}
