// Package world provides creature management
package world

import (
	"math"

	"voxelgame/internal/generation/entity"
	vmath "voxelgame/pkg/math"

	"github.com/go-gl/mathgl/mgl32"
)

// CreatureManager manages creatures in the world
type CreatureManager struct {
	// Creature generator
	generator *entity.Generator

	// Active creatures
	creatures []*entity.Creature

	// Configuration
	maxCreatures      int
	spawnRadius       float32
	despawnRadius     float32
	creaturesPerChunk int

	// RNG
	rng *vmath.SeededRNG
}

// NewCreatureManager creates a new creature manager
func NewCreatureManager(seed int64) *CreatureManager {
	return &CreatureManager{
		generator:         entity.NewGenerator(seed),
		creatures:         make([]*entity.Creature, 0, 100),
		maxCreatures:      50,
		spawnRadius:       50,
		despawnRadius:     80,
		creaturesPerChunk: 2,
		rng:               vmath.NewSeededRNG(seed + 5000),
	}
}

// Update updates all creatures and handles spawning/despawning
func (cm *CreatureManager) Update(dt float32, playerPos mgl32.Vec3, getBiome func(x, z int) string, getHeight func(x, z int) int) {
	// Update existing creatures
	for i := len(cm.creatures) - 1; i >= 0; i-- {
		creature := cm.creatures[i]

		// Update AI
		creature.Update(dt, playerPos)

		// Ground creature to terrain
		cm.groundCreature(creature, getHeight)

		// Check despawn distance
		dx := creature.Position.X() - playerPos.X()
		dz := creature.Position.Z() - playerPos.Z()
		dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))

		if dist > cm.despawnRadius {
			// Remove creature
			cm.creatures = append(cm.creatures[:i], cm.creatures[i+1:]...)
		}
	}

	// Try to spawn new creatures
	if len(cm.creatures) < cm.maxCreatures && cm.rng.Next() < 0.02 {
		cm.trySpawn(playerPos, getBiome, getHeight)
	}
}

// groundCreature ensures the creature is properly positioned on the terrain
func (cm *CreatureManager) groundCreature(c *entity.Creature, getHeight func(x, z int) int) {
	// Get terrain height at creature position
	terrainHeight := float32(getHeight(int(c.Position.X()), int(c.Position.Z())))
	c.GroundY = terrainHeight

	switch c.Template {
	case entity.TemplateFlying:
		// Flying creatures hover above terrain
		hoverHeight := terrainHeight + 3.0 + float32(math.Sin(float64(c.AnimationTime)*2.0))*0.5
		// Smoothly interpolate to hover height
		c.Position[1] += (hoverHeight - c.Position.Y()) * 0.1

	case entity.TemplateSlime:
		// Slimes can bounce, apply gravity when above ground
		if c.Position.Y() > terrainHeight+0.01 {
			c.Velocity[1] -= 15.0 * 0.016 // Gravity
		}
		// Clamp to ground
		if c.Position.Y() < terrainHeight {
			c.Position[1] = terrainHeight
			c.Velocity[1] = 0
		}

	case entity.TemplateFish:
		// Fish stay at water level or terrain if no water
		waterLevel := float32(12) // Sea level
		if terrainHeight < waterLevel {
			c.Position[1] = waterLevel - 1.0 + float32(math.Sin(float64(c.AnimationTime)*3.0))*0.3
		} else {
			// On land - die or flop (for now just stay on ground)
			c.Position[1] = terrainHeight
		}

	default:
		// Quadrupeds, bipeds, spiders - walk on ground
		// Apply gravity
		if c.Position.Y() > terrainHeight+0.01 {
			c.Velocity[1] -= 20.0 * 0.016 // Gravity
		}
		// Clamp to ground
		if c.Position.Y() < terrainHeight {
			c.Position[1] = terrainHeight
			c.Velocity[1] = 0
		}
	}
}

// trySpawn attempts to spawn a new creature
func (cm *CreatureManager) trySpawn(playerPos mgl32.Vec3, getBiome func(x, z int) string, getHeight func(x, z int) int) {
	// Random position around player
	angle := cm.rng.NextFloat(0, 2*math.Pi)
	dist := cm.rng.NextFloat(float64(cm.spawnRadius*0.5), float64(cm.spawnRadius))

	spawnX := playerPos.X() + float32(math.Cos(angle)*dist)
	spawnZ := playerPos.Z() + float32(math.Sin(angle)*dist)

	// Get biome and height
	biome := getBiome(int(spawnX), int(spawnZ))
	height := getHeight(int(spawnX), int(spawnZ))

	if height <= 12 { // Below sea level
		return
	}

	spawnY := float32(height) + 1.0
	spawnPos := mgl32.Vec3{spawnX, spawnY, spawnZ}

	// Choose template based on biome
	template := cm.chooseTemplate(biome)

	// Create creature
	size := float32(cm.rng.NextFloat(0.6, 1.4))
	creature := cm.generator.Create(template, biome, spawnPos, size)

	cm.creatures = append(cm.creatures, creature)
}

// chooseTemplate selects a creature template based on biome
func (cm *CreatureManager) chooseTemplate(biome string) entity.CreatureTemplate {
	switch biome {
	case "desert":
		// Spiders and slimes in desert
		if cm.rng.Next() > 0.7 {
			return entity.TemplateSpider
		}
		return entity.TemplateSlime

	case "snow":
		// Quadrupeds in snow
		return entity.TemplateQuadruped

	case "forest":
		// Mix in forest
		r := cm.rng.Next()
		if r < 0.4 {
			return entity.TemplateQuadruped
		} else if r < 0.7 {
			return entity.TemplateBiped
		}
		return entity.TemplateFlying

	case "mountains":
		// Flying and quadrupeds
		if cm.rng.Next() > 0.5 {
			return entity.TemplateFlying
		}
		return entity.TemplateQuadruped

	default: // plains
		r := cm.rng.Next()
		if r < 0.5 {
			return entity.TemplateQuadruped
		} else if r < 0.8 {
			return entity.TemplateSlime
		}
		return entity.TemplateBiped
	}
}

// GetCreatures returns all active creatures
func (cm *CreatureManager) GetCreatures() []*entity.Creature {
	return cm.creatures
}

// GetCreatureCount returns the number of active creatures
func (cm *CreatureManager) GetCreatureCount() int {
	return len(cm.creatures)
}

// Clear removes all creatures
func (cm *CreatureManager) Clear() {
	cm.creatures = cm.creatures[:0]
}

// SpawnCreature manually spawns a creature at a position
func (cm *CreatureManager) SpawnCreature(template entity.CreatureTemplate, biome string, pos mgl32.Vec3) *entity.Creature {
	creature := cm.generator.Create(template, biome, pos, 0)
	cm.creatures = append(cm.creatures, creature)
	return creature
}
