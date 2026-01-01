// Package terrain provides procedural terrain generation
package terrain

import (
	"voxelgame/internal/core/block"
	"voxelgame/internal/core/chunk"
	"voxelgame/internal/core/noise"
	vmath "voxelgame/pkg/math"
)

// World generation constants
const (
	SeaLevel          = 12
	TerrainBaseHeight = 20
	TerrainAmplitude  = 30
)

// Biome represents a terrain biome with its properties
type Biome struct {
	Name       string
	Surface    block.Type
	Subsurface block.Type
	HeightMod  float64
	HasWater   bool
	WaterType  block.Type
	HasTrees   bool
	TreeChance float64
	TreeType   string
	HasFlowers bool
	HasCactus  bool
}

// Predefined biomes
var (
	BiomePlains = Biome{
		Name:       "plains",
		Surface:    block.Grass,
		Subsurface: block.Dirt,
		HeightMod:  0.5,
		HasWater:   true,
		HasTrees:   true,
		TreeChance: 0.01,
		TreeType:   "oak",
		HasFlowers: true,
	}

	BiomeDesert = Biome{
		Name:       "desert",
		Surface:    block.Sand,
		Subsurface: block.Sand,
		HeightMod:  0.3,
		HasWater:   true,
		HasCactus:  true,
	}

	BiomeSnow = Biome{
		Name:       "snow",
		Surface:    block.Snow,
		Subsurface: block.Dirt,
		HeightMod:  0.7,
		HasWater:   true,
		WaterType:  block.Ice,
		HasTrees:   true,
		TreeChance: 0.02,
		TreeType:   "spruce",
	}

	BiomeForest = Biome{
		Name:       "forest",
		Surface:    block.Grass,
		Subsurface: block.Dirt,
		HeightMod:  0.6,
		HasWater:   true,
		HasTrees:   true,
		TreeChance: 0.08,
		TreeType:   "oak",
		HasFlowers: true,
	}

	BiomeMountains = Biome{
		Name:       "mountains",
		Surface:    block.Stone,
		Subsurface: block.Stone,
		HeightMod:  1.5,
		HasWater:   true,
		HasTrees:   true,
		TreeChance: 0.005,
		TreeType:   "spruce",
	}
)

// Generator generates procedural terrain
type Generator struct {
	seed int64
	rng  *vmath.SeededRNG

	// Configuration
	Config GeneratorConfig

	// Noise generators
	heightNoise *noise.SimplexNoise
	biomeNoise  *noise.SimplexNoise
	caveNoise   *noise.SimplexNoise
	detailNoise *noise.SimplexNoise

	// FBM configurations
	heightFBM *noise.FBM
	biomeFBM  *noise.FBM
	caveFBM   *noise.FBM
}

// GeneratorConfig holds terrain generation settings
type GeneratorConfig struct {
	SeaLevel         int
	TerrainAmplitude float32
	TreeDensity      float32
	CaveFrequency    float32
}

// DefaultConfig returns default generation config
func DefaultConfig() GeneratorConfig {
	return GeneratorConfig{
		SeaLevel:         12,
		TerrainAmplitude: 30.0,
		TreeDensity:      0.05,
		CaveFrequency:    0.6,
	}
}

// NewGenerator creates a new terrain generator with the given seed
func NewGenerator(seed int64) *Generator {
	g := &Generator{
		seed:        seed,
		rng:         vmath.NewSeededRNG(seed),
		Config:      DefaultConfig(), // Use defaults initially
		heightNoise: noise.NewSimplexNoise(seed),
		biomeNoise:  noise.NewSimplexNoise(seed + 1000),
		caveNoise:   noise.NewSimplexNoise(seed + 2000),
		detailNoise: noise.NewSimplexNoise(seed + 3000),
	}

	g.heightFBM = noise.NewFBM(noise.FBMConfig{
		Octaves:     6,
		Lacunarity:  2.0,
		Persistence: 0.5,
		Scale:       0.005,
	})

	g.biomeFBM = noise.NewFBM(noise.FBMConfig{
		Octaves:     4,
		Lacunarity:  2.0,
		Persistence: 0.5,
		Scale:       0.002,
	})

	g.caveFBM = noise.NewFBM(noise.FBMConfig{
		Octaves:     3,
		Lacunarity:  2.0,
		Persistence: 0.5,
		Scale:       0.05,
	})

	return g
}

// GenerateChunk generates terrain for a chunk
func (g *Generator) GenerateChunk(c *chunk.Chunk) {
	startX := int(c.CX) * chunk.Size
	startZ := int(c.CZ) * chunk.Size

	// First pass: base terrain
	for lx := 0; lx < chunk.Size; lx++ {
		for lz := 0; lz < chunk.Size; lz++ {
			wx := startX + lx
			wz := startZ + lz

			g.generateColumn(c, lx, lz, wx, wz)
		}
	}

	// Second pass: structures (trees, cacti)
	g.generateStructures(c, startX, startZ)

	// Third pass: decorations (flowers, grass)
	g.generateDecorations(c, startX, startZ)

	// Fourth pass: waterfalls & lakes
	g.generateWaterfalls(c, startX, startZ)

	// Fifth pass: dungeons (inside caves)
	g.generateDungeons(c, startX, startZ)

	// Sixth pass: surface campfires
	g.generateCampfires(c, startX, startZ)

	c.IsGenerated = true
}

// generateColumn generates a vertical column of blocks
func (g *Generator) generateColumn(c *chunk.Chunk, lx, lz, wx, wz int) {
	biome := g.getBiome(wx, wz)
	baseHeight := g.getTerrainHeight(wx, wz, biome)

	// Update height map
	c.HeightMap[lx+lz*chunk.Size] = uint8(baseHeight)

	for y := 0; y < chunk.Height; y++ {
		var blockType block.Type = block.Air

		if y == 0 {
			// Bedrock
			blockType = block.Bedrock
		} else if y < baseHeight-4 {
			// Underground
			blockType = g.getUndergroundBlock(wx, y, wz, biome)
		} else if y < baseHeight {
			// Subsurface layer
			blockType = biome.Subsurface
		} else if y == baseHeight {
			// Surface
			blockType = g.getSurfaceBlock(y, biome)
		} else if y < g.Config.SeaLevel && biome.HasWater {
			// Water
			blockType = block.Water
		}

		if blockType != block.Air {
			c.SetBlock(lx, y, lz, blockType)
		}
	}
}

// getBiome determines the biome at a world position
func (g *Generator) getBiome(wx, wz int) Biome {
	temperature := g.biomeFBM.Sample2D(g.biomeNoise, float64(wx), float64(wz))
	humidity := g.biomeFBM.Sample2D(g.biomeNoise, float64(wx)+5000, float64(wz)+5000)

	if temperature > 0.3 {
		if humidity < -0.2 {
			return BiomeDesert
		}
		return BiomePlains
	} else if temperature < -0.3 {
		return BiomeSnow
	} else {
		if humidity > 0.2 {
			return BiomeForest
		}
		return BiomeMountains
	}
}

// getTerrainHeight calculates terrain height at a position
func (g *Generator) getTerrainHeight(wx, wz int, biome Biome) int {
	height := float64(20) // Base height hardcoded replacing TerrainBaseHeight constant

	// FBM for general terrain
	fbmValue := g.heightFBM.Sample2D(g.heightNoise, float64(wx), float64(wz))

	// Smooth height modulation based on temperature
	// This prevents sharp cliffs between biomes by blending the HeightMod
	temperature := g.biomeFBM.Sample2D(g.biomeNoise, float64(wx), float64(wz))

	// Base height mod (Plains/Forest)
	heightMod := 0.5

	// Adjust based on temperature (extreme temps = higher/rougher terrain)
	if temperature < -0.3 {
		// Snow/Mountains
		factor := (temperature + 0.3) * -5.0 // 0..1
		heightMod = 0.5 + factor*1.0         // 0.5 .. 1.5
	} else if temperature > 0.3 {
		// Desert/Plains - flattening
		factor := (temperature - 0.3) * 5.0
		heightMod = 0.5 - factor*0.2 // 0.5 .. 0.3
	}

	// Apply smoothed modulation and configured amplitude
	height += fbmValue * float64(g.Config.TerrainAmplitude) * heightMod

	// High frequency detail
	detail := g.detailNoise.Noise2D(float64(wx)*0.1, float64(wz)*0.1) * 2
	height += detail

	// Mountains get ridged noise (blended)
	if biome.Name == "mountains" {
		ridged := g.heightFBM.Ridged2D(g.heightNoise, float64(wx)*2, float64(wz)*2)
		height += ridged * 20
	} else if temperature < -0.2 && temperature > -0.4 {
		// Transition to mountains - add some ridged noise seamlessly
		ridged := g.heightFBM.Ridged2D(g.heightNoise, float64(wx)*2, float64(wz)*2)
		blend := (temperature + 0.2) / -0.2 // 0 at -0.2, 1 at -0.4
		// Wait, math: if temp = -0.3. top = -0.1. bottom = -0.2. result = 0.5.
		// if temp = -0.4. top = -0.2. bottom = -0.2. result = 1.0.
		blend = absDouble(blend) // Clamp?
		if blend > 1.0 {
			blend = 1.0
		}
		height += ridged * 10 * blend
	}

	result := int(height)
	if result < 1 {
		result = 1
	}
	if result > chunk.Height-10 {
		result = chunk.Height - 10
	}
	return result
}

func absDouble(n float64) float64 {
	if n < 0 {
		return -n
	}
	return n
}

// getUndergroundBlock determines block type underground
func (g *Generator) getUndergroundBlock(wx, y, wz int, biome Biome) block.Type {
	// Caves
	caveValue := g.caveFBM.Sample3D(g.caveNoise, float64(wx), float64(y), float64(wz))
	if caveValue > float64(g.Config.CaveFrequency) && y > 5 {
		// Lava at the bottom of deep caves
		if y < 10 && caveValue > float64(g.Config.CaveFrequency)+0.05 {
			return block.Lava
		}
		return block.Air
	}

	// Ores
	oreChance := g.detailNoise.Noise3D(float64(wx)*0.2, float64(y)*0.2, float64(wz)*0.2)

	if y < 15 && oreChance > 0.85 {
		return block.DiamondOre
	} else if y < 30 && oreChance > 0.8 {
		return block.GoldOre
	} else if y < 45 && oreChance > 0.75 {
		return block.IronOre
	} else if oreChance > 0.7 {
		return block.CoalOre
	}

	return block.Stone
}

// getSurfaceBlock determines the surface block
func (g *Generator) getSurfaceBlock(height int, biome Biome) block.Type {
	if height <= g.Config.SeaLevel+2 && biome.Name != "desert" {
		return block.Sand
	}
	return biome.Surface
}

// generateStructures generates trees and cacti
func (g *Generator) generateStructures(c *chunk.Chunk, startX, startZ int) {
	chunkRng := vmath.NewSeededRNG(g.seed + int64(c.CX)*1000 + int64(c.CZ))

	for lx := 2; lx < chunk.Size-2; lx++ {
		for lz := 2; lz < chunk.Size-2; lz++ {
			wx := startX + lx
			wz := startZ + lz
			height := c.GetHeight(lx, lz)

			if height <= g.Config.SeaLevel {
				continue
			}

			biome := g.getBiome(wx, wz)

			// Trees - scale chance by tree density config
			// Base chance is for density 0.05. Scaling: config / 0.05
			densityMultiplier := g.Config.TreeDensity / 0.05

			if biome.HasTrees && chunkRng.Next() < biome.TreeChance*float64(densityMultiplier) {
				g.generateTree(c, lx, height+1, lz, biome.TreeType, chunkRng)
			}

			// Cacti
			if biome.HasCactus && chunkRng.Next() < 0.005 {
				g.generateCactus(c, lx, height+1, lz, chunkRng)
			}
		}
	}
}

// generateTree generates a tree at the given position
func (g *Generator) generateTree(c *chunk.Chunk, lx, ly, lz int, treeType string, rng *vmath.SeededRNG) {
	height := 4 + rng.NextInt(0, 2)

	var logType, leafType block.Type
	switch treeType {
	case "birch":
		logType = block.BirchLog
		leafType = block.BirchLeaves
	case "spruce":
		logType = block.SpruceLog
		leafType = block.SpruceLeaves
	default: // oak
		logType = block.OakLog
		leafType = block.OakLeaves
	}

	// Trunk
	for i := 0; i < height; i++ {
		if ly+i < chunk.Height {
			c.SetBlock(lx, ly+i, lz, logType)
		}
	}

	// Leaves
	leafStart := height - 2
	for dy := leafStart; dy <= height+1; dy++ {
		radius := 2
		if dy == height+1 {
			radius = 1
		}

		for dx := -radius; dx <= radius; dx++ {
			for dz := -radius; dz <= radius; dz++ {
				if abs(dx)+abs(dz) <= radius+1 {
					nlx := lx + dx
					nlz := lz + dz
					nly := ly + dy

					if nlx >= 0 && nlx < chunk.Size && nlz >= 0 && nlz < chunk.Size && nly < chunk.Height {
						if c.GetBlock(nlx, nly, nlz) == block.Air {
							c.SetBlock(nlx, nly, nlz, leafType)
						}
					}
				}
			}
		}
	}
}

// generateCactus generates a cactus
func (g *Generator) generateCactus(c *chunk.Chunk, lx, ly, lz int, rng *vmath.SeededRNG) {
	height := 2 + rng.NextInt(0, 2)

	for i := 0; i < height; i++ {
		if ly+i < chunk.Height {
			c.SetBlock(lx, ly+i, lz, block.Cactus)
		}
	}
}

// SetConfig updates the generator configuration
func (g *Generator) SetConfig(config GeneratorConfig) {
	g.Config = config
}

// generateDecorations generates flowers and tall grass
func (g *Generator) generateDecorations(c *chunk.Chunk, startX, startZ int) {
	chunkRng := vmath.NewSeededRNG(g.seed + int64(c.CX)*2000 + int64(c.CZ))

	for lx := 0; lx < chunk.Size; lx++ {
		for lz := 0; lz < chunk.Size; lz++ {
			wx := startX + lx
			wz := startZ + lz
			height := c.GetHeight(lx, lz)

			if height <= g.Config.SeaLevel {
				continue
			}

			biome := g.getBiome(wx, wz)
			surfaceBlock := c.GetBlock(lx, height, lz)

			if surfaceBlock != block.Grass {
				continue
			}

			if biome.HasFlowers {
				// Tall grass
				if chunkRng.Next() < 0.15 {
					c.SetBlock(lx, height+1, lz, block.TallGrass)
				} else if chunkRng.Next() < 0.02 {
					// Flowers
					var flowerType block.Type
					if chunkRng.Next() > 0.5 {
						flowerType = block.FlowerRed
					} else {
						flowerType = block.FlowerYellow
					}
					c.SetBlock(lx, height+1, lz, flowerType)
				}
			}

			// Mushrooms (rare)
			if chunkRng.Next() < 0.005 {
				var mushroom block.Type
				if chunkRng.Next() > 0.5 {
					mushroom = block.MushroomRed
				} else {
					mushroom = block.MushroomBrown
				}
				c.SetBlock(lx, height+1, lz, mushroom)
			}
		}
	}
}

// generateWaterfalls generates waterfalls in mountain biomes
func (g *Generator) generateWaterfalls(c *chunk.Chunk, startX, startZ int) {
	chunkRng := vmath.NewSeededRNG(g.seed + int64(c.CX)*3000 + int64(c.CZ))

	// 15% chance per chunk
	if chunkRng.Next() > 0.15 {
		return
	}

	for lx := 3; lx < chunk.Size-3; lx++ {
		for lz := 3; lz < chunk.Size-3; lz++ {
			wx := startX + lx
			wz := startZ + lz
			biome := g.getBiome(wx, wz)

			if biome.Name != "mountains" {
				continue
			}

			height := c.GetHeight(lx, lz)
			if height < 35 {
				continue
			}

			// Check for cliffs
			directions := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

			for _, dir := range directions {
				if lx+dir[0]*2 < 0 || lx+dir[0]*2 >= chunk.Size ||
					lz+dir[1]*2 < 0 || lz+dir[1]*2 >= chunk.Size {
					continue
				}

				neighborHeight := c.GetHeight(lx+dir[0]*2, lz+dir[1]*2)
				heightDiff := height - neighborHeight

				if heightDiff >= 8 && chunkRng.Next() < 0.3 {
					// Place water source at top
					c.SetBlock(lx, height, lz, block.Water)

					// Create cascade
					currentY := height - 1
					currentX := lx + dir[0]
					currentZ := lz + dir[1]

					for currentY > neighborHeight && currentY > g.Config.SeaLevel {
						if currentX >= 0 && currentX < chunk.Size &&
							currentZ >= 0 && currentZ < chunk.Size {
							if c.GetBlock(currentX, currentY, currentZ) == block.Air {
								c.SetBlock(currentX, currentY, currentZ, block.Water)
							}
						}
						currentY--
					}

					// Create a lake at the base
					g.generateLake(c, currentX, neighborHeight, currentZ, 3, block.Water)

					// Only one waterfall per chunk
					return
				}
			}
		}
	}
}

// generateLake creates a small circular pool
func (g *Generator) generateLake(c *chunk.Chunk, lx, ly, lz, radius int, liquid block.Type) {
	for dx := -radius; dx <= radius; dx++ {
		for dz := -radius; dz <= radius; dz++ {
			if dx*dx+dz*dz <= radius*radius {
				nx, nz := lx+dx, lz+dz
				if nx >= 0 && nx < chunk.Size && nz >= 0 && nz < chunk.Size {
					// Carve out a bit of the shore if needed, and fill with liquid
					for dy := -1; dy <= 0; dy++ {
						ny := ly + dy
						if ny > 0 && ny < chunk.Height {
							c.SetBlock(nx, ny, nz, liquid)
						}
					}
				}
			}
		}
	}
}

// generateDungeons places stone brick structures in large cave pockets
func (g *Generator) generateDungeons(c *chunk.Chunk, startX, startZ int) {
	chunkRng := vmath.NewSeededRNG(g.seed + int64(c.CX)*4000 + int64(c.CZ))

	// 5% chance per chunk for a dungeon
	if chunkRng.Next() > 0.05 {
		return
	}

	// Try to find a suitable cave location
	for attempt := 0; attempt < 10; attempt++ {
		lx := chunkRng.NextInt(4, chunk.Size-4)
		lz := chunkRng.NextInt(4, chunk.Size-4)
		ly := chunkRng.NextInt(10, 30)

		// Check if it's in a cave
		if c.GetBlock(lx, ly, lz) == block.Air {
			g.buildDungeonRoom(c, lx, ly, lz, chunkRng)
			return
		}
	}
}

func (g *Generator) buildDungeonRoom(c *chunk.Chunk, x, y, z int, rng *vmath.SeededRNG) {
	width := rng.NextInt(5, 8)
	height := rng.NextInt(4, 6)
	depth := rng.NextInt(5, 8)

	for dx := -width / 2; dx <= width/2; dx++ {
		for dy := 0; dy < height; dy++ {
			for dz := -depth / 2; dz <= depth/2; dz++ {
				nx, ny, nz := x+dx, y+dy, z+dz
				if nx < 0 || nx >= chunk.Size || ny < 0 || ny >= chunk.Height || nz < 0 || nz >= chunk.Size {
					continue
				}

				isWall := dx == -width/2 || dx == width/2 || dy == 0 || dy == height-1 || dz == -depth/2 || dz == depth/2

				if isWall {
					brick := block.StoneBrick
					if rng.Next() < 0.2 {
						brick = block.MossyStoneBrick
					}
					c.SetBlock(nx, ny, nz, brick)
				} else {
					c.SetBlock(nx, ny, nz, block.Air)
				}
			}
		}
	}
}

// generateCampfires places fogueiras on the surface
func (g *Generator) generateCampfires(c *chunk.Chunk, startX, startZ int) {
	chunkRng := vmath.NewSeededRNG(g.seed + int64(c.CX)*5000 + int64(c.CZ))

	// Very rare: 2% chance per chunk
	if chunkRng.Next() > 0.02 {
		return
	}

	lx := chunkRng.NextInt(2, chunk.Size-2)
	lz := chunkRng.NextInt(2, chunk.Size-2)
	wx := startX + lx
	wz := startZ + lz
	biome := g.getBiome(wx, wz)

	if biome.Name == "plains" || biome.Name == "forest" {
		height := c.GetHeight(lx, lz)
		if height > g.Config.SeaLevel {
			c.SetBlock(lx, height+1, lz, block.Campfire)
		}
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// GetBiomeName returns the biome name at world coordinates
func (g *Generator) GetBiomeName(wx, wz int) string {
	return g.getBiome(wx, wz).Name
}
