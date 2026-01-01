// Package entity provides procedural creature generation
package entity

import (
	"math"

	"voxelgame/internal/core/block"
	vmath "voxelgame/pkg/math"

	"github.com/go-gl/mathgl/mgl32"
)

// CreatureTemplate defines the type of creature
type CreatureTemplate string

const (
	TemplateQuadruped CreatureTemplate = "quadruped" // 4 legs (land animals)
	TemplateBiped     CreatureTemplate = "biped"     // 2 legs (humanoids)
	TemplateFlying    CreatureTemplate = "flying"    // Wings (birds)
	TemplateSlime     CreatureTemplate = "slime"     // Blob
	TemplateFish      CreatureTemplate = "fish"      // Aquatic
	TemplateSpider    CreatureTemplate = "spider"    // 8 legs
)

// AllTemplates contains all creature templates
var AllTemplates = []CreatureTemplate{
	TemplateQuadruped, TemplateBiped, TemplateFlying,
	TemplateSlime, TemplateFish, TemplateSpider,
}

// Behavior defines creature behaviors
type Behavior string

const (
	BehaviorWander Behavior = "wander"
	BehaviorFollow Behavior = "follow"
	BehaviorFlee   Behavior = "flee"
	BehaviorIdle   Behavior = "idle"
	BehaviorJump   Behavior = "jump"
	BehaviorSwim   Behavior = "swim"
)

// BiomeColors defines creature color palettes by biome
var BiomeColors = map[string][][3]float32{
	"plains":    {{0.55, 0.27, 0.07}, {0.85, 0.65, 0.13}, {0.96, 0.87, 0.70}, {0.82, 0.71, 0.55}, {0.63, 0.32, 0.18}},
	"forest":    {{0.13, 0.55, 0.13}, {0.55, 0.27, 0.07}, {0.18, 0.55, 0.34}, {0.42, 0.56, 0.14}, {0.33, 0.42, 0.18}},
	"desert":    {{0.88, 0.75, 0.56}, {0.87, 0.72, 0.53}, {0.96, 0.64, 0.38}, {0.82, 0.41, 0.12}, {0.80, 0.52, 0.25}},
	"snow":      {{1.0, 1.0, 1.0}, {0.94, 0.97, 1.0}, {0.90, 0.90, 0.98}, {0.69, 0.77, 0.87}, {0.47, 0.53, 0.60}},
	"mountains": {{0.41, 0.41, 0.41}, {0.50, 0.50, 0.50}, {0.66, 0.66, 0.66}, {0.18, 0.31, 0.31}, {0.44, 0.50, 0.56}},
}

// BodyPart represents a part of a creature's body
type BodyPart struct {
	Type   string // "torso", "head", "leg", "arm", "wing", "tail", etc.
	Size   mgl32.Vec3
	Offset mgl32.Vec3
}

// CreatureStats contains creature statistics
type CreatureStats struct {
	Health    int
	MaxHealth int
	Speed     float32
	JumpForce float32
	Damage    int
	Hostile   bool
}

// Creature represents a procedurally generated creature
type Creature struct {
	ID       int
	Template CreatureTemplate
	Size     float32
	Biome    string

	// Body
	BodyParts []BodyPart
	Height    float32
	Width     float32
	Depth     float32

	// Appearance
	PrimaryColor   [3]float32
	SecondaryColor [3]float32
	AccentColor    [3]float32

	// Behavior
	Behaviors []Behavior
	State     string

	// Position and movement
	Position mgl32.Vec3
	Velocity mgl32.Vec3
	Rotation float32

	// AI state
	Target *mgl32.Vec3
	Timer  float32

	// Stats
	Stats CreatureStats

	// Inventory
	HeldItem block.Type

	// Animation state
	AnimationTime float32 // Continuous time for animations
	WalkPhase     float32 // 0-2π cycle for walk animation
	IsMoving      bool    // Whether creature is currently moving
	GroundY       float32 // Current ground level Y position
}

// Generator creates procedural creatures
type Generator struct {
	seed       int64
	rng        *vmath.SeededRNG
	creatureID int
}

// NewGenerator creates a new creature generator
func NewGenerator(seed int64) *Generator {
	return &Generator{
		seed: seed,
		rng:  vmath.NewSeededRNG(seed),
	}
}

// Create creates a new procedural creature
func (g *Generator) Create(template CreatureTemplate, biome string, position mgl32.Vec3, size float32) *Creature {
	if size <= 0 {
		size = float32(g.rng.NextFloat(0.5, 2.0))
	}

	c := &Creature{
		ID:       g.creatureID,
		Template: template,
		Size:     size,
		Biome:    biome,
		Position: position,
		State:    "idle",
	}
	g.creatureID++

	// Generate body
	c.BodyParts = g.generateBody(template, size)
	c.Height = size
	c.Width = size * 0.6
	c.Depth = size * 0.8

	// Generate colors
	c.PrimaryColor, c.SecondaryColor, c.AccentColor = g.generateColors(biome)

	// Generate behaviors
	c.Behaviors = g.generateBehaviors(template)

	// Generate stats
	c.Stats = g.generateStats(template, size)

	return c
}

// CreateRandom creates a random creature for the given biome
func (g *Generator) CreateRandom(biome string, position mgl32.Vec3) *Creature {
	template := vmath.Choose(g.rng, AllTemplates)
	return g.Create(template, biome, position, 0)
}

func (g *Generator) generateBody(template CreatureTemplate, size float32) []BodyPart {
	var parts []BodyPart

	switch template {
	case TemplateQuadruped:
		parts = []BodyPart{
			{Type: "torso", Size: mgl32.Vec3{0.8, 0.5, 1.2}, Offset: mgl32.Vec3{0, 0.5, 0}},
			{Type: "head", Size: mgl32.Vec3{0.4, 0.4, 0.5}, Offset: mgl32.Vec3{0, 0.7, 0.6}},
			{Type: "leg", Size: mgl32.Vec3{0.15, 0.5, 0.15}, Offset: mgl32.Vec3{-0.3, 0, 0.4}},
			{Type: "leg", Size: mgl32.Vec3{0.15, 0.5, 0.15}, Offset: mgl32.Vec3{0.3, 0, 0.4}},
			{Type: "leg", Size: mgl32.Vec3{0.15, 0.5, 0.15}, Offset: mgl32.Vec3{-0.3, 0, -0.4}},
			{Type: "leg", Size: mgl32.Vec3{0.15, 0.5, 0.15}, Offset: mgl32.Vec3{0.3, 0, -0.4}},
			{Type: "tail", Size: mgl32.Vec3{0.1, 0.1, 0.4}, Offset: mgl32.Vec3{0, 0.5, -0.8}},
		}

	case TemplateBiped:
		parts = []BodyPart{
			{Type: "torso", Size: mgl32.Vec3{0.5, 0.7, 0.3}, Offset: mgl32.Vec3{0, 0.8, 0}},
			{Type: "head", Size: mgl32.Vec3{0.35, 0.35, 0.35}, Offset: mgl32.Vec3{0, 1.35, 0}},
			{Type: "arm", Size: mgl32.Vec3{0.15, 0.6, 0.15}, Offset: mgl32.Vec3{-0.35, 0.9, 0}},
			{Type: "arm", Size: mgl32.Vec3{0.15, 0.6, 0.15}, Offset: mgl32.Vec3{0.35, 0.9, 0}},
			{Type: "leg", Size: mgl32.Vec3{0.2, 0.7, 0.2}, Offset: mgl32.Vec3{-0.15, 0, 0}},
			{Type: "leg", Size: mgl32.Vec3{0.2, 0.7, 0.2}, Offset: mgl32.Vec3{0.15, 0, 0}},
		}

	case TemplateFlying:
		parts = []BodyPart{
			{Type: "body", Size: mgl32.Vec3{0.3, 0.3, 0.5}, Offset: mgl32.Vec3{0, 0, 0}},
			{Type: "head", Size: mgl32.Vec3{0.2, 0.2, 0.25}, Offset: mgl32.Vec3{0, 0.1, 0.3}},
			{Type: "wing", Size: mgl32.Vec3{0.8, 0.05, 0.3}, Offset: mgl32.Vec3{-0.5, 0.1, 0}},
			{Type: "wing", Size: mgl32.Vec3{0.8, 0.05, 0.3}, Offset: mgl32.Vec3{0.5, 0.1, 0}},
			{Type: "tail", Size: mgl32.Vec3{0.15, 0.05, 0.4}, Offset: mgl32.Vec3{0, 0, -0.4}},
		}

	case TemplateSlime:
		blobSize := float32(0.5 + g.rng.NextFloat(0, 0.5))
		parts = []BodyPart{
			{Type: "blob", Size: mgl32.Vec3{blobSize, blobSize * 0.8, blobSize}, Offset: mgl32.Vec3{0, blobSize * 0.4, 0}},
		}

	case TemplateFish:
		parts = []BodyPart{
			{Type: "body", Size: mgl32.Vec3{0.2, 0.3, 0.6}, Offset: mgl32.Vec3{0, 0, 0}},
			{Type: "fin", Size: mgl32.Vec3{0.3, 0.2, 0.1}, Offset: mgl32.Vec3{0, 0.2, 0}},
			{Type: "tail", Size: mgl32.Vec3{0.05, 0.25, 0.2}, Offset: mgl32.Vec3{0, 0, -0.35}},
		}

	case TemplateSpider:
		parts = []BodyPart{
			{Type: "abdomen", Size: mgl32.Vec3{0.5, 0.4, 0.6}, Offset: mgl32.Vec3{0, 0.3, -0.3}},
			{Type: "thorax", Size: mgl32.Vec3{0.3, 0.25, 0.3}, Offset: mgl32.Vec3{0, 0.25, 0.2}},
			{Type: "head", Size: mgl32.Vec3{0.2, 0.2, 0.2}, Offset: mgl32.Vec3{0, 0.25, 0.4}},
		}
		// Add 8 legs
		for i := 0; i < 8; i++ {
			side := float32(-1)
			if i%2 == 0 {
				side = 1
			}
			zOffset := (float32(i)/2.0 - 1.5) * 0.15
			parts = append(parts, BodyPart{
				Type:   "leg",
				Size:   mgl32.Vec3{0.3, 0.05, 0.05},
				Offset: mgl32.Vec3{side * 0.3, 0.2, zOffset},
			})
		}
	}

	// Scale all parts by size
	for i := range parts {
		parts[i].Size = parts[i].Size.Mul(size)
		parts[i].Offset = parts[i].Offset.Mul(size)
	}

	return parts
}

func (g *Generator) generateColors(biome string) (primary, secondary, accent [3]float32) {
	palette, ok := BiomeColors[biome]
	if !ok {
		palette = BiomeColors["plains"]
	}

	primary = vmath.Choose(g.rng, palette)
	secondary = vmath.Choose(g.rng, palette)

	// Lighten primary for accent
	accent = [3]float32{
		min32(1.0, primary[0]+0.3),
		min32(1.0, primary[1]+0.3),
		min32(1.0, primary[2]+0.3),
	}

	return
}

func (g *Generator) generateBehaviors(template CreatureTemplate) []Behavior {
	behaviors := []Behavior{BehaviorIdle, BehaviorWander}

	switch template {
	case TemplateQuadruped:
		behaviors = append(behaviors, BehaviorFlee)
	case TemplateBiped:
		behaviors = append(behaviors, BehaviorFollow)
	case TemplateFlying:
		behaviors = append(behaviors, BehaviorFlee)
	case TemplateSlime:
		behaviors = append(behaviors, BehaviorJump)
	case TemplateFish:
		behaviors = append(behaviors, BehaviorSwim)
	case TemplateSpider:
		behaviors = append(behaviors, BehaviorFollow)
	}

	return behaviors
}

func (g *Generator) generateStats(template CreatureTemplate, size float32) CreatureStats {
	baseHealth := 10
	baseSpeed := float32(2.0)

	stats := CreatureStats{
		Health:    int(float32(baseHealth) * size),
		MaxHealth: int(float32(baseHealth) * size),
		Speed:     baseSpeed / size, // Smaller = faster
		JumpForce: 5.0,
		Damage:    int(2 * size),
		Hostile:   false,
	}

	if template == TemplateSlime {
		stats.JumpForce = 8.0
	}
	if template == TemplateSpider {
		stats.Hostile = true
	}

	return stats
}

// Update updates creature behavior
func (c *Creature) Update(dt float32, playerPos mgl32.Vec3) {
	c.Timer += dt
	c.AnimationTime += dt

	switch c.State {
	case "idle":
		if c.Timer > 2+float32(math.Mod(float64(c.Timer), 3)) {
			c.State = "wander"
			c.Timer = 0
			// Set random target nearby
			target := mgl32.Vec3{
				c.Position.X() + float32(math.Cos(float64(c.Timer)*10))*10,
				c.Position.Y(),
				c.Position.Z() + float32(math.Sin(float64(c.Timer)*10))*10,
			}
			c.Target = &target
		}

	case "wander":
		if c.Target != nil {
			dx := c.Target.X() - c.Position.X()
			dz := c.Target.Z() - c.Position.Z()
			dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))

			if dist < 0.5 || c.Timer > 5 {
				c.State = "idle"
				c.Timer = 0
				c.Target = nil
			} else {
				c.Velocity[0] = (dx / dist) * c.Stats.Speed
				c.Velocity[2] = (dz / dist) * c.Stats.Speed
				c.Rotation = float32(math.Atan2(float64(dx), float64(dz)))
			}
		}

	case "flee":
		dx := c.Position.X() - playerPos.X()
		dz := c.Position.Z() - playerPos.Z()
		dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))

		if dist > 15 {
			c.State = "idle"
			c.Timer = 0
		} else {
			c.Velocity[0] = (dx / dist) * c.Stats.Speed * 1.5
			c.Velocity[2] = (dz / dist) * c.Stats.Speed * 1.5
			c.Rotation = float32(math.Atan2(float64(-dx), float64(-dz)))
		}
	}

	// Slimes jump periodically
	if c.Template == TemplateSlime && c.Timer > 1 {
		if c.Position.Y() < c.GroundY+0.1 {
			c.Velocity[1] = c.Stats.JumpForce
			c.Timer = 0
		}
	}

	// Apply velocity
	c.Position = c.Position.Add(c.Velocity.Mul(dt))

	// Update animation state
	speed := float32(math.Sqrt(float64(c.Velocity[0]*c.Velocity[0] + c.Velocity[2]*c.Velocity[2])))
	c.IsMoving = speed > 0.1

	if c.IsMoving {
		// Walk phase cycles at speed proportional to movement
		c.WalkPhase += dt * speed * 3.0
		if c.WalkPhase > 2*math.Pi {
			c.WalkPhase -= 2 * math.Pi
		}
	}

	// Dampen velocity
	c.Velocity[0] *= 0.9
	c.Velocity[2] *= 0.9
}

func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
