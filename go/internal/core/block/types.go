// Package block defines all block types and their properties
package block

// Type represents a block type identifier
type Type uint8

// Block type constants - matching the JavaScript version
const (
	Air Type = iota
	Grass
	Dirt
	Stone
	Wood
	Leaves
	Sand
	Water
	Snow
	Ice
	Clay
	Gravel
	Cobblestone
	Bedrock
	CoalOre
	IronOre
	GoldOre
	DiamondOre
	Cactus
	DeadBush
	FlowerRed
	FlowerYellow
	MushroomRed
	MushroomBrown
	TallGrass
	OakLog
	BirchLog
	SpruceLog
	OakLeaves
	BirchLeaves
	SpruceLeaves
	Glass
	Brick
	BlockTypeCount // Total number of block types
)

// Definition contains all properties for a block type
type Definition struct {
	Name           string
	Solid          bool
	Transparent    bool
	Collidable     bool
	Color          [3]float32 // RGB normalized [0-1]
	BreakTime      float32
	Liquid         bool
	Opacity        float32
	Gravity        bool // Falls like sand
	Damages        bool // Damages player on contact
	Indestructible bool
	Emissive       float32 // Light emission
}

// String returns the block type name
func (t Type) String() string {
	if def, ok := Registry[t]; ok {
		return def.Name
	}
	return "Unknown"
}

// IsAir returns true if the block is air
func (t Type) IsAir() bool {
	return t == Air
}

// IsSolid returns true if the block is solid
func (t Type) IsSolid() bool {
	if def, ok := Registry[t]; ok {
		return def.Solid
	}
	return false
}

// IsTransparent returns true if the block is transparent
func (t Type) IsTransparent() bool {
	if def, ok := Registry[t]; ok {
		return def.Transparent
	}
	return true
}

// IsCollidable returns true if the block is collidable
func (t Type) IsCollidable() bool {
	if def, ok := Registry[t]; ok {
		return def.Collidable
	}
	return false
}

// IsLiquid returns true if the block is a liquid
func (t Type) IsLiquid() bool {
	if def, ok := Registry[t]; ok {
		return def.Liquid
	}
	return false
}

// GetColor returns the block color as RGB
func (t Type) GetColor() [3]float32 {
	if def, ok := Registry[t]; ok {
		return def.Color
	}
	return [3]float32{1, 0, 1} // Magenta for unknown
}
