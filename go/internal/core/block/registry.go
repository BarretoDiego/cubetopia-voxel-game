// Package block contains the block registry with all block definitions
package block

// hexToRGB converts hex color string to normalized RGB
func hexToRGB(hex string) [3]float32 {
	if len(hex) < 7 || hex[0] != '#' {
		return [3]float32{1, 0, 1} // Magenta for invalid
	}

	var r, g, b int
	_, _ = parseHexByte(hex[1:3], &r)
	_, _ = parseHexByte(hex[3:5], &g)
	_, _ = parseHexByte(hex[5:7], &b)

	return [3]float32{
		float32(r) / 255.0,
		float32(g) / 255.0,
		float32(b) / 255.0,
	}
}

func parseHexByte(s string, result *int) (bool, error) {
	val := 0
	for _, c := range s {
		val *= 16
		switch {
		case c >= '0' && c <= '9':
			val += int(c - '0')
		case c >= 'a' && c <= 'f':
			val += int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			val += int(c-'A') + 10
		}
	}
	*result = val
	return true, nil
}

// Registry contains all block definitions
var Registry = map[Type]Definition{
	Air: {
		Name:        "Ar",
		Solid:       false,
		Transparent: true,
		Collidable:  false,
		Color:       [3]float32{0, 0, 0},
	},
	Grass: {
		Name:          "Grama",
		Solid:         true,
		Transparent:   false,
		Collidable:    true,
		Color:         hexToRGB("#567d46"),
		BreakTime:     0.5,
		TextureTop:    1, // Grass Top
		TextureSide:   2, // Grass Side
		TextureBottom: 0, // Dirt
	},
	Dirt: {
		Name:          "Terra",
		Solid:         true,
		Transparent:   false,
		Collidable:    true,
		Color:         hexToRGB("#8b6914"),
		BreakTime:     0.5,
		Material:      MaterialStandard,
		TextureTop:    0,
		TextureSide:   0,
		TextureBottom: 0,
	},
	Stone: {
		Name:          "Pedra",
		Solid:         true,
		Transparent:   false,
		Collidable:    true,
		Color:         hexToRGB("#7a7a7a"),
		BreakTime:     4.0,
		Material:      MaterialStone,
		TextureTop:    3,
		TextureSide:   3,
		TextureBottom: 3,
	},
	Wood: {
		Name:          "Madeira",
		Solid:         true,
		Transparent:   false,
		Collidable:    true,
		Color:         hexToRGB("#8b5a2b"),
		BreakTime:     1.5,
		TextureTop:    4,
		TextureSide:   4,
		TextureBottom: 4,
	},
	Leaves: {
		Name:          "Folhas",
		Solid:         true,
		Transparent:   true,
		Collidable:    true,
		Color:         hexToRGB("#228b22"),
		BreakTime:     0.2,
		Material:      MaterialFoliage,
		TextureTop:    5,
		TextureSide:   5,
		TextureBottom: 5,
	},
	Sand: {
		Name:          "Areia",
		Solid:         true,
		Transparent:   false,
		Collidable:    true,
		Color:         hexToRGB("#e0c090"),
		BreakTime:     0.5,
		Gravity:       true,
		TextureTop:    8,
		TextureSide:   8,
		TextureBottom: 8,
	},
	Water: {
		Name:          "Água",
		Solid:         false,
		Transparent:   true,
		Collidable:    false,
		Color:         hexToRGB("#3498db"),
		Opacity:       0.6,
		Liquid:        true,
		Material:      MaterialLiquid,
		TextureTop:    6,
		TextureSide:   6,
		TextureBottom: 6,
	},
	Snow: {
		Name:          "Neve",
		Solid:         true,
		Transparent:   false,
		Collidable:    true,
		Color:         hexToRGB("#f0f0f0"),
		BreakTime:     0.3,
		TextureTop:    9,
		TextureSide:   9,
		TextureBottom: 9,
	},
	Ice: {
		Name:          "Gelo",
		Solid:         true,
		Transparent:   true,
		Collidable:    true,
		Color:         hexToRGB("#a5f2f3"),
		Opacity:       0.8,
		BreakTime:     0.5,
		Material:      MaterialGlass,
		TextureTop:    7,
		TextureSide:   7,
		TextureBottom: 7,
	},
	Clay: {
		Name:        "Argila",
		Solid:       true,
		Transparent: false,
		Collidable:  true,
		Color:       hexToRGB("#9fa4ad"),
		BreakTime:   0.6,
	},
	Gravel: {
		Name:        "Cascalho",
		Solid:       true,
		Transparent: false,
		Collidable:  true,
		Color:       hexToRGB("#808080"),
		BreakTime:   0.6,
		Gravity:     true,
	},
	Cobblestone: {
		Name:        "Paralelepípedos",
		Solid:       true,
		Transparent: false,
		Collidable:  true,
		Color:       hexToRGB("#5a5a5a"),
		BreakTime:   2.5,
	},
	Bedrock: {
		Name:           "Rocha-mãe",
		Solid:          true,
		Transparent:    false,
		Collidable:     true,
		Color:          hexToRGB("#1a1a1a"),
		Indestructible: true,
	},
	CoalOre: {
		Name:        "Carvão",
		Solid:       true,
		Transparent: false,
		Collidable:  true,
		Color:       hexToRGB("#2a2a2a"),
		BreakTime:   4.0,
	},
	IronOre: {
		Name:        "Ferro",
		Solid:       true,
		Transparent: false,
		Collidable:  true,
		Color:       hexToRGB("#b8945f"),
		BreakTime:   4.0,
	},
	GoldOre: {
		Name:        "Ouro",
		Solid:       true,
		Transparent: false,
		Collidable:  true,
		Color:       hexToRGB("#fcee4b"),
		BreakTime:   4.0,
	},
	DiamondOre: {
		Name:        "Diamante",
		Solid:       true,
		Transparent: false,
		Collidable:  true,
		Color:       hexToRGB("#4aedd9"),
		BreakTime:   8.0,
		Emissive:    0.2,
	},
	Cactus: {
		Name:        "Cacto",
		Solid:       true,
		Transparent: false,
		Collidable:  true,
		Color:       hexToRGB("#0b5d1e"),
		BreakTime:   0.4,
		Damages:     true,
	},
	DeadBush: {
		Name:        "Arbusto Seco",
		Solid:       false,
		Transparent: true,
		Collidable:  false,
		Color:       hexToRGB("#8b7355"),
		BreakTime:   0,
	},
	FlowerRed: {
		Name:          "Flor Vermelha",
		Solid:         false,
		Transparent:   true,
		Collidable:    false,
		Color:         hexToRGB("#ff4444"),
		BreakTime:     0,
		Material:      MaterialFoliage,
		HasCustomMesh: true,
	},
	FlowerYellow: {
		Name:          "Flor Amarela",
		Solid:         false,
		Transparent:   true,
		Collidable:    false,
		Color:         hexToRGB("#ffff44"),
		BreakTime:     0,
		Material:      MaterialFoliage,
		HasCustomMesh: true,
	},
	MushroomRed: {
		Name:        "Cogumelo Vermelho",
		Solid:       false,
		Transparent: true,
		Collidable:  false,
		Color:       hexToRGB("#ff0000"),
		BreakTime:   0,
	},
	MushroomBrown: {
		Name:        "Cogumelo Marrom",
		Solid:       false,
		Transparent: true,
		Collidable:  false,
		Color:       hexToRGB("#8b4513"),
		BreakTime:   0,
	},
	TallGrass: {
		Name:          "Grama Alta",
		Solid:         false,
		Transparent:   true,
		Collidable:    false,
		Color:         hexToRGB("#4a7023"),
		BreakTime:     0,
		Material:      MaterialFoliage,
		HasCustomMesh: true,
	},
	OakLog: {
		Name:        "Tronco de Carvalho",
		Solid:       true,
		Transparent: false,
		Collidable:  true,
		Color:       hexToRGB("#6b4423"),
		BreakTime:   1.5,
	},
	BirchLog: {
		Name:        "Tronco de Bétula",
		Solid:       true,
		Transparent: false,
		Collidable:  true,
		Color:       hexToRGB("#d5c4a1"),
		BreakTime:   1.5,
	},
	SpruceLog: {
		Name:        "Tronco de Pinheiro",
		Solid:       true,
		Transparent: false,
		Collidable:  true,
		Color:       hexToRGB("#3e2723"),
		BreakTime:   1.5,
	},
	OakLeaves: {
		Name:        "Folhas de Carvalho",
		Solid:       true,
		Transparent: true,
		Collidable:  true,
		Color:       hexToRGB("#228b22"),
		BreakTime:   0.2,
		Material:    MaterialFoliage,
	},
	BirchLeaves: {
		Name:        "Folhas de Bétula",
		Solid:       true,
		Transparent: true,
		Collidable:  true,
		Color:       hexToRGB("#80c622"),
		BreakTime:   0.2,
	},
	SpruceLeaves: {
		Name:        "Folhas de Pinheiro",
		Solid:       true,
		Transparent: true,
		Collidable:  true,
		Color:       hexToRGB("#1a472a"),
		BreakTime:   0.2,
	},
	Glass: {
		Name:          "Vidro",
		Solid:         true,
		Transparent:   true,
		Collidable:    true,
		Color:         hexToRGB("#c8dbe0"),
		Opacity:       0.3,
		BreakTime:     0.3,
		Material:      MaterialGlass,
		TextureTop:    10,
		TextureSide:   10,
		TextureBottom: 10,
	},
	Brick: {
		Name:        "Tijolo",
		Solid:       true,
		Transparent: false,
		Collidable:  true,
		Color:       hexToRGB("#b75a3c"),
		BreakTime:   2.5,
	},
	Pickaxe: {
		Name:        "Picareta",
		Solid:       false,
		Transparent: true,
		Collidable:  false,
		Color:       hexToRGB("#3498db"),
	},
	Axe: {
		Name:        "Machado",
		Solid:       false,
		Transparent: true,
		Collidable:  false,
		Color:       hexToRGB("#8b4513"),
	},
	Sword: {
		Name:        "Espada",
		Solid:       false,
		Transparent: true,
		Collidable:  false,
		Color:       hexToRGB("#bdc3c7"),
	},
	Shovel: {
		Name:        "Pá",
		Solid:       false,
		Transparent: true,
		Collidable:  false,
		Color:       hexToRGB("#95a5a6"),
	},
	Lava: {
		Name:          "Lava",
		Solid:         false,
		Transparent:   false,
		Collidable:    false,
		Color:         hexToRGB("#ff4500"),
		Liquid:        true,
		Material:      MaterialLiquid,
		Emissive:      1.0,
		TextureTop:    11,
		TextureSide:   11,
		TextureBottom: 11,
		Damages:       true,
	},
	Campfire: {
		Name:          "Fogueira",
		Solid:         false,
		Transparent:   true,
		Collidable:    false,
		Color:         hexToRGB("#ff8c00"),
		Emissive:      0.8,
		TextureTop:    12,
		TextureSide:   12,
		TextureBottom: 12,
		HasCustomMesh: true,
		Material:      MaterialFoliage,
	},
	StoneBrick: {
		Name:          "Tijolo de Pedra",
		Solid:         true,
		Transparent:   false,
		Collidable:    true,
		Color:         hexToRGB("#a9a9a9"),
		BreakTime:     1.5,
		TextureTop:    13,
		TextureSide:   13,
		TextureBottom: 13,
		Material:      MaterialStone,
	},
	MossyStoneBrick: {
		Name:          "Tijolo de Pedra Musgoso",
		Solid:         true,
		Transparent:   false,
		Collidable:    true,
		Color:         hexToRGB("#8b9467"),
		BreakTime:     1.5,
		TextureTop:    14,
		TextureSide:   14,
		TextureBottom: 14,
		Material:      MaterialStone,
	},
}

// GetDefinition returns the definition for a block type
func GetDefinition(t Type) Definition {
	if def, ok := Registry[t]; ok {
		return def
	}
	return Registry[Air]
}

// GetAllPlaceableBlocks returns all blocks that can be placed by the player
func GetAllPlaceableBlocks() []Type {
	placeable := make([]Type, 0, BlockTypeCount)
	for t := Type(1); t < BlockTypeCount; t++ { // Skip Air
		def := Registry[t]
		// Include solid blocks, water, and tools
		isTool := t == Pickaxe || t == Axe || t == Sword || t == Shovel
		if def.Solid || t == Water || isTool {
			placeable = append(placeable, t)
		}
	}
	return placeable
}
