package render

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"

	"voxelgame/assets"
)

// ParticleType represents different atmospheric particle types
type ParticleType int

const (
	ParticleTypeDefault ParticleType = iota
	ParticleTypeDust                 // Desert/plains floating dust motes
	ParticleTypeSpore                // Forest spores/pollen
	ParticleTypeFirefly              // Glowing fireflies at night
	ParticleTypeSnow                 // Falling snowflakes
	ParticleTypeRain                 // Rain droplets
	ParticleTypeAsh                  // Volcanic ash
	ParticleTypeMist                 // Swamp/water mist
	ParticleTypeFire                 // Fire/flame particles
	ParticleTypeSmoke                // Smoke particles rising
	ParticleTypeBubble               // Underwater bubbles
	ParticleTypeFish                 // Fish silhouettes in water
	ParticleTypeSpark                // Sparks from fire
)

// Particle represents a single particle
type Particle struct {
	Position mgl32.Vec3
	Velocity mgl32.Vec3
	Color    mgl32.Vec4
	Life     float32
	Size     float32
	Type     ParticleType
}

// ParticleSystem manages particles
type ParticleSystem struct {
	particles    []Particle
	maxParticles int

	// Rendering
	vao         uint32
	vbo         uint32 // Vertex buffer (quad)
	instanceVBO uint32 // Instance buffer (positions/colors)
	shader      *Shader
	texture     uint32 // Particle texture (e.g. circle/gleam)
}

// NewParticleSystem creates a new particle system
func NewParticleSystem(maxParticles int) (*ParticleSystem, error) {
	ps := &ParticleSystem{
		particles:    make([]Particle, 0, maxParticles),
		maxParticles: maxParticles,
	}

	// Initialize rendering assets
	if err := ps.initGL(); err != nil {
		return nil, err
	}

	return ps, nil
}

// initGL sets up the particle VAO/VBO and shader
func (ps *ParticleSystem) initGL() error {
	// 1. Create Shader from embedded assets
	vSource, err := assets.ReadFile("shaders/particle.vert")
	if err == nil {
		fSource, fErr := assets.ReadFile("shaders/particle.frag")
		if fErr == nil {
			ps.shader, err = NewShader(string(vSource), string(fSource))
			if err != nil {
				fmt.Printf("Warning: Failed to compile particle shader: %v\n", err)
			}
		} else {
			fmt.Printf("Warning: Failed to read particle fragment shader: %v\n", fErr)
		}
	} else {
		fmt.Printf("Warning: Failed to read particle vertex shader: %v\n", err)
	}

	// 2. VAO/VBO Setup
	// Simple quad vertices
	quadVertices := []float32{
		// Pos        // UV
		-0.5, -0.5, 0.0, 0.0, 0.0,
		0.5, -0.5, 0.0, 1.0, 0.0,
		0.5, 0.5, 0.0, 1.0, 1.0,
		-0.5, 0.5, 0.0, 0.0, 1.0,
	}
	// Indices? Just Arrays for now or EBO. Let's use Triangle Fan or 2 triangles.
	// Let's use array drawing 6 vertices or EBO.
	indices := []uint32{0, 1, 2, 0, 2, 3}

	gl.GenVertexArrays(1, &ps.vao)
	gl.GenBuffers(1, &ps.vbo)

	gl.BindVertexArray(ps.vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, ps.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(quadVertices)*4, gl.Ptr(quadVertices), gl.STATIC_DRAW)

	// Mesh Attributes (Pos, UV)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 5*4, 0)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, 5*4, 3*4)

	// Instance VBO (Position, Color, Size)
	gl.GenBuffers(1, &ps.instanceVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, ps.instanceVBO)
	// Empty init
	gl.BufferData(gl.ARRAY_BUFFER, ps.maxParticles*(3+4+1)*4, nil, gl.STREAM_DRAW)

	// Instance Attributes start at 2
	stride := int32((3 + 4 + 1) * 4)

	// Instance Pos (Location 2)
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointerWithOffset(2, 3, gl.FLOAT, false, stride, 0)
	gl.VertexAttribDivisor(2, 1) // Per instance

	// Instance Color (Location 3)
	gl.EnableVertexAttribArray(3)
	gl.VertexAttribPointerWithOffset(3, 4, gl.FLOAT, false, stride, 3*4)
	gl.VertexAttribDivisor(3, 1)

	// Instance Size (Location 4)
	gl.EnableVertexAttribArray(4)
	gl.VertexAttribPointerWithOffset(4, 1, gl.FLOAT, false, stride, 7*4)
	gl.VertexAttribDivisor(4, 1)

	// Use EBO
	var ebo uint32
	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	gl.BindVertexArray(0)

	return nil
}

// Update updates all particles
func (ps *ParticleSystem) Update(dt float32) {
	aliveCount := 0
	for i := 0; i < len(ps.particles); i++ {
		p := &ps.particles[i]
		p.Life -= dt

		if p.Life > 0 {
			// Update Physics based on particle type
			switch p.Type {
			case ParticleTypeDust, ParticleTypeSpore:
				// Floating particles - gentle drift, no gravity
				// Add slight wave motion
				wave := float32(math.Sin(float64(p.Life*3))) * 0.1
				p.Velocity[1] = wave * 0.5
				p.Position = p.Position.Add(p.Velocity.Mul(dt))

			case ParticleTypeFirefly:
				// Fireflies - random wandering motion
				wander := mgl32.Vec3{
					(rand.Float32() - 0.5) * 2,
					(rand.Float32() - 0.5) * 1,
					(rand.Float32() - 0.5) * 2,
				}
				p.Velocity = p.Velocity.Add(wander.Mul(dt * 5))
				// Damping
				p.Velocity = p.Velocity.Mul(0.98)
				p.Position = p.Position.Add(p.Velocity.Mul(dt))
				// Pulsing glow effect by modifying alpha
				pulse := 0.5 + 0.5*float32(math.Sin(float64(p.Life*8)))
				p.Color[3] = pulse * 0.9

			case ParticleTypeSnow:
				// Snow - slow falling with slight horizontal drift
				p.Velocity[1] = -1.5 // Constant slow fall
				// Add wind effect
				p.Velocity[0] += (rand.Float32() - 0.5) * 0.1
				p.Velocity[2] += (rand.Float32() - 0.5) * 0.1
				// Damping horizontal movement
				p.Velocity[0] *= 0.98
				p.Velocity[2] *= 0.98
				p.Position = p.Position.Add(p.Velocity.Mul(dt))

			case ParticleTypeRain:
				// Rain - fast falling
				p.Velocity[1] = -15.0
				p.Position = p.Position.Add(p.Velocity.Mul(dt))

			case ParticleTypeMist:
				// Mist - very slow floating upward
				p.Velocity[1] = 0.3
				p.Velocity[0] += (rand.Float32() - 0.5) * 0.05
				p.Velocity[2] += (rand.Float32() - 0.5) * 0.05
				p.Position = p.Position.Add(p.Velocity.Mul(dt))
				// Fade out over time
				p.Color[3] = p.Life / 5.0 * 0.4

			case ParticleTypeFire:
				// Fire - rises quickly, flickers, fades from yellow to red
				p.Velocity[1] = 3.0 + rand.Float32()*2.0 // Rising
				p.Velocity[0] += (rand.Float32() - 0.5) * 0.5
				p.Velocity[2] += (rand.Float32() - 0.5) * 0.5
				p.Position = p.Position.Add(p.Velocity.Mul(dt))
				// Color transition: yellow -> orange -> red as life decreases
				lifeRatio := p.Life / 2.0 // Assuming ~2s life
				if lifeRatio > 0.5 {
					p.Color = mgl32.Vec4{1.0, 0.9, 0.3, 0.9} // Yellow
				} else if lifeRatio > 0.2 {
					p.Color = mgl32.Vec4{1.0, 0.5, 0.1, 0.8} // Orange
				} else {
					p.Color = mgl32.Vec4{0.8, 0.2, 0.0, p.Life / 0.5} // Red, fading
				}
				// Shrink as it rises
				p.Size *= 0.99

			case ParticleTypeSmoke:
				// Smoke - rises slowly, expands, fades to transparent
				p.Velocity[1] = 1.5 + rand.Float32()*0.5
				p.Velocity[0] += (rand.Float32() - 0.5) * 0.1
				p.Velocity[2] += (rand.Float32() - 0.5) * 0.1
				p.Position = p.Position.Add(p.Velocity.Mul(dt))
				// Expand over time
				p.Size *= 1.01
				// Fade out
				p.Color[3] = p.Life / 4.0 * 0.5

			case ParticleTypeSpark:
				// Sparks - gravity affected, bright, short life
				p.Velocity[1] -= 15.0 * dt // Gravity
				p.Position = p.Position.Add(p.Velocity.Mul(dt))
				// Flickering brightness
				flicker := 0.7 + rand.Float32()*0.3
				p.Color[3] = flicker * p.Life / 0.5

			case ParticleTypeBubble:
				// Bubbles - rise slowly, wobble side to side
				p.Velocity[1] = 2.0 + rand.Float32()*0.5
				wobble := float32(math.Sin(float64(p.Life*10))) * 0.3
				p.Velocity[0] = wobble
				p.Velocity[2] = float32(math.Cos(float64(p.Life*8))) * 0.2
				p.Position = p.Position.Add(p.Velocity.Mul(dt))
				// Fade out near surface
				p.Color[3] = 0.6

			case ParticleTypeFish:
				// Fish - swim in random directions, turn occasionally
				if rand.Float32() < 0.02 { // 2% chance to change direction
					p.Velocity[0] = (rand.Float32() - 0.5) * 4
					p.Velocity[1] = (rand.Float32() - 0.5) * 1
					p.Velocity[2] = (rand.Float32() - 0.5) * 4
				}
				p.Position = p.Position.Add(p.Velocity.Mul(dt))
				// Keep fish visible
				p.Color[3] = 0.8

			default:
				// Default behavior - gravity affected
				p.Velocity = p.Velocity.Add(mgl32.Vec3{0, -9.8 * dt, 0})
				p.Position = p.Position.Add(p.Velocity.Mul(dt))
			}

			// Keep alive
			ps.particles[aliveCount] = *p
			aliveCount++
		}
	}
	ps.particles = ps.particles[:aliveCount]
}

// Render renders particles
func (ps *ParticleSystem) Render(view, projection mgl32.Mat4) {
	if len(ps.particles) == 0 || ps.shader == nil {
		return
	}

	ps.shader.Use()
	ps.shader.SetMat4("uView", view)
	ps.shader.SetMat4("uProjection", projection)
	// ps.shader.SetInt("uTexture", 0)

	// Prepare instance data and upload
	// Flatten data: PosX, PosY, PosZ, R, G, B, A, Size
	data := make([]float32, 0, len(ps.particles)*8)
	for _, p := range ps.particles {
		data = append(data, p.Position.X(), p.Position.Y(), p.Position.Z())
		data = append(data, p.Color.X(), p.Color.Y(), p.Color.Z(), p.Color.W())
		data = append(data, p.Size)
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, ps.instanceVBO)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(data)*4, gl.Ptr(data))

	gl.BindVertexArray(ps.vao)
	// Draw Instanced
	gl.DrawElementsInstanced(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil, int32(len(ps.particles)))
	gl.BindVertexArray(0)
}

// Emit spawns a new particle
func (ps *ParticleSystem) Emit(pos, vel mgl32.Vec3, color mgl32.Vec4, life, size float32) {
	if len(ps.particles) >= ps.maxParticles {
		return // Pool full
	}
	ps.particles = append(ps.particles, Particle{
		Position: pos,
		Velocity: vel,
		Color:    color,
		Life:     life,
		Size:     size,
		Type:     ParticleTypeDefault,
	})
}

// EmitTyped spawns a new particle with a specific type
func (ps *ParticleSystem) EmitTyped(pos, vel mgl32.Vec3, color mgl32.Vec4, life, size float32, pType ParticleType) {
	if len(ps.particles) >= ps.maxParticles {
		return // Pool full
	}
	ps.particles = append(ps.particles, Particle{
		Position: pos,
		Velocity: vel,
		Color:    color,
		Life:     life,
		Size:     size,
		Type:     pType,
	})
}

// EmitExplosion creates a burst of particles
func (ps *ParticleSystem) EmitExplosion(pos mgl32.Vec3, count int, color mgl32.Vec4) {
	for i := 0; i < count; i++ {
		vel := mgl32.Vec3{
			(rand.Float32() - 0.5) * 5,
			(rand.Float32() * 5) + 2,
			(rand.Float32() - 0.5) * 5,
		}
		life := 0.5 + rand.Float32()
		size := 0.1 + rand.Float32()*0.2
		ps.Emit(pos, vel, color, life, size)
	}
}

// UpdateAtmospheric spawns ambient particles based on biome and player position
// Call this every frame to maintain atmospheric particle density
func (ps *ParticleSystem) UpdateAtmospheric(playerPos mgl32.Vec3, biome string, timeOfDay float32, dt float32) {
	// Configuration
	const (
		spawnRadius     = 25.0 // How far around player to spawn particles
		spawnHeight     = 15.0 // Vertical range (above player)
		particlesPerSec = 25   // Base particles to spawn per second
		maxAtmospheric  = 600  // Max atmospheric particles at once
	)

	// Count current atmospheric particles (rough estimate based on particle pool usage)
	currentAtmospheric := len(ps.particles)
	if currentAtmospheric >= maxAtmospheric {
		return // Cap atmospheric particles
	}

	// Determine particle types based on biome
	var particlesToSpawn []struct {
		pType    ParticleType
		color    mgl32.Vec4
		size     float32
		life     float32
		velocity mgl32.Vec3
		rate     float32 // spawn rate multiplier
	}

	isNight := timeOfDay > 0.7 || timeOfDay < 0.2

	switch biome {
	case "desert":
		// Dust motes floating in the air
		particlesToSpawn = append(particlesToSpawn, struct {
			pType    ParticleType
			color    mgl32.Vec4
			size     float32
			life     float32
			velocity mgl32.Vec3
			rate     float32
		}{
			pType:    ParticleTypeDust,
			color:    mgl32.Vec4{0.9, 0.85, 0.7, 0.6}, // Sandy color (brighter)
			size:     0.4,
			life:     8.0,
			velocity: mgl32.Vec3{0.5, 0.1, 0.3},
			rate:     2.0,
		})

	case "forest":
		// Forest spores/pollen
		particlesToSpawn = append(particlesToSpawn, struct {
			pType    ParticleType
			color    mgl32.Vec4
			size     float32
			life     float32
			velocity mgl32.Vec3
			rate     float32
		}{
			pType:    ParticleTypeSpore,
			color:    mgl32.Vec4{0.8, 0.9, 0.6, 0.7}, // Yellow-green pollen (brighter)
			size:     0.35,
			life:     10.0,
			velocity: mgl32.Vec3{0.2, 0.05, 0.2},
			rate:     1.5,
		})
		// Fireflies at night
		if isNight {
			particlesToSpawn = append(particlesToSpawn, struct {
				pType    ParticleType
				color    mgl32.Vec4
				size     float32
				life     float32
				velocity mgl32.Vec3
				rate     float32
			}{
				pType:    ParticleTypeFirefly,
				color:    mgl32.Vec4{0.9, 1.0, 0.4, 1.0}, // Glowing yellow-green (bright)
				size:     0.5,
				life:     15.0,
				velocity: mgl32.Vec3{0.5, 0, 0.5},
				rate:     1.0,
			})
		}

	case "snow", "tundra", "taiga":
		// Snowflakes falling
		particlesToSpawn = append(particlesToSpawn, struct {
			pType    ParticleType
			color    mgl32.Vec4
			size     float32
			life     float32
			velocity mgl32.Vec3
			rate     float32
		}{
			pType:    ParticleTypeSnow,
			color:    mgl32.Vec4{1.0, 1.0, 1.0, 0.9}, // White (brighter)
			size:     0.5,
			life:     12.0,
			velocity: mgl32.Vec3{0.3, -1.0, 0.3},
			rate:     3.0,
		})

	case "swamp":
		// Mist rising
		particlesToSpawn = append(particlesToSpawn, struct {
			pType    ParticleType
			color    mgl32.Vec4
			size     float32
			life     float32
			velocity mgl32.Vec3
			rate     float32
		}{
			pType:    ParticleTypeMist,
			color:    mgl32.Vec4{0.7, 0.8, 0.7, 0.5}, // Greenish mist (brighter)
			size:     2.0,
			life:     5.0,
			velocity: mgl32.Vec3{0.1, 0.3, 0.1},
			rate:     1.2,
		})
		// Fireflies at night in swamps too
		if isNight {
			particlesToSpawn = append(particlesToSpawn, struct {
				pType    ParticleType
				color    mgl32.Vec4
				size     float32
				life     float32
				velocity mgl32.Vec3
				rate     float32
			}{
				pType:    ParticleTypeFirefly,
				color:    mgl32.Vec4{0.6, 1.0, 0.5, 1.0},
				size:     0.5,
				life:     12.0,
				velocity: mgl32.Vec3{0.3, 0, 0.3},
				rate:     1.0,
			})
		}

	case "plains", "grassland":
		// Light dust/pollen
		particlesToSpawn = append(particlesToSpawn, struct {
			pType    ParticleType
			color    mgl32.Vec4
			size     float32
			life     float32
			velocity mgl32.Vec3
			rate     float32
		}{
			pType:    ParticleTypeDust,
			color:    mgl32.Vec4{1.0, 1.0, 0.9, 0.5}, // Light particles (brighter)
			size:     0.3,
			life:     10.0,
			velocity: mgl32.Vec3{0.4, 0.05, 0.4},
			rate:     1.0,
		})

	default:
		// Generic floating dust for any biome
		particlesToSpawn = append(particlesToSpawn, struct {
			pType    ParticleType
			color    mgl32.Vec4
			size     float32
			life     float32
			velocity mgl32.Vec3
			rate     float32
		}{
			pType:    ParticleTypeDust,
			color:    mgl32.Vec4{0.9, 0.9, 0.9, 0.4},
			size:     0.25,
			life:     8.0,
			velocity: mgl32.Vec3{0.2, 0.02, 0.2},
			rate:     0.8,
		})
	}

	// Spawn particles based on dt and rate
	for _, p := range particlesToSpawn {
		spawnChance := float32(particlesPerSec) * p.rate * dt
		numToSpawn := int(spawnChance)
		if rand.Float32() < (spawnChance - float32(numToSpawn)) {
			numToSpawn++
		}

		for i := 0; i < numToSpawn; i++ {
			// Random position within sphere around player
			angle := rand.Float32() * 2 * math.Pi
			distance := rand.Float32() * spawnRadius
			height := rand.Float32() * spawnHeight

			pos := mgl32.Vec3{
				playerPos.X() + float32(math.Cos(float64(angle)))*distance,
				playerPos.Y() + height - spawnHeight/2,
				playerPos.Z() + float32(math.Sin(float64(angle)))*distance,
			}

			// Randomize velocity slightly
			vel := mgl32.Vec3{
				p.velocity.X() * (0.5 + rand.Float32()),
				p.velocity.Y() * (0.5 + rand.Float32()),
				p.velocity.Z() * (0.5 + rand.Float32()),
			}

			// Randomize life slightly
			life := p.life * (0.8 + rand.Float32()*0.4)

			// Randomize size slightly
			size := p.size * (0.8 + rand.Float32()*0.4)

			// Randomize color alpha slightly
			color := p.color
			color[3] *= 0.7 + rand.Float32()*0.3

			ps.EmitTyped(pos, vel, color, life, size, p.pType)
		}
	}
}

// GetParticleCount returns the current number of active particles
func (ps *ParticleSystem) GetParticleCount() int {
	return len(ps.particles)
}

// EmitCampfire emits fire, smoke, and spark particles at a campfire location
// Call this every frame for each campfire in the scene
func (ps *ParticleSystem) EmitCampfire(pos mgl32.Vec3, dt float32) {
	// Fire particles - many per frame
	fireRate := 20.0 * dt
	numFire := int(fireRate)
	if rand.Float32() < float32(fireRate)-float32(numFire) {
		numFire++
	}

	for i := 0; i < numFire; i++ {
		firePos := mgl32.Vec3{
			pos.X() + (rand.Float32()-0.5)*0.5,
			pos.Y() + rand.Float32()*0.3,
			pos.Z() + (rand.Float32()-0.5)*0.5,
		}
		vel := mgl32.Vec3{
			(rand.Float32() - 0.5) * 0.5,
			2.0 + rand.Float32()*2.0,
			(rand.Float32() - 0.5) * 0.5,
		}
		color := mgl32.Vec4{1.0, 0.8, 0.2, 0.9} // Bright yellow-orange
		life := 0.5 + rand.Float32()*1.5
		size := 0.3 + rand.Float32()*0.3
		ps.EmitTyped(firePos, vel, color, life, size, ParticleTypeFire)
	}

	// Smoke particles - fewer, above fire
	smokeRate := 5.0 * dt
	numSmoke := int(smokeRate)
	if rand.Float32() < float32(smokeRate)-float32(numSmoke) {
		numSmoke++
	}

	for i := 0; i < numSmoke; i++ {
		smokePos := mgl32.Vec3{
			pos.X() + (rand.Float32()-0.5)*0.3,
			pos.Y() + 1.0 + rand.Float32()*0.5,
			pos.Z() + (rand.Float32()-0.5)*0.3,
		}
		vel := mgl32.Vec3{
			(rand.Float32() - 0.5) * 0.2,
			1.0 + rand.Float32()*0.5,
			(rand.Float32() - 0.5) * 0.2,
		}
		color := mgl32.Vec4{0.3, 0.3, 0.3, 0.4} // Gray smoke
		life := 3.0 + rand.Float32()*2.0
		size := 0.4 + rand.Float32()*0.3
		ps.EmitTyped(smokePos, vel, color, life, size, ParticleTypeSmoke)
	}

	// Sparks - occasional
	if rand.Float32() < 0.1*dt*60 { // ~10% chance per frame at 60fps
		sparkPos := mgl32.Vec3{
			pos.X() + (rand.Float32()-0.5)*0.3,
			pos.Y() + 0.5,
			pos.Z() + (rand.Float32()-0.5)*0.3,
		}
		vel := mgl32.Vec3{
			(rand.Float32() - 0.5) * 3,
			3.0 + rand.Float32()*4.0,
			(rand.Float32() - 0.5) * 3,
		}
		color := mgl32.Vec4{1.0, 0.7, 0.2, 1.0} // Bright orange
		life := 0.3 + rand.Float32()*0.4
		size := 0.1 + rand.Float32()*0.1
		ps.EmitTyped(sparkPos, vel, color, life, size, ParticleTypeSpark)
	}
}

// UpdateWaterParticles emits bubbles and fish particles in water near the player
func (ps *ParticleSystem) UpdateWaterParticles(playerPos mgl32.Vec3, isUnderwater bool, waterSurfaceY float32, dt float32) {
	if isUnderwater {
		// Bubbles rising from player
		if rand.Float32() < 0.3*dt*60 {
			bubblePos := mgl32.Vec3{
				playerPos.X() + (rand.Float32()-0.5)*1.0,
				playerPos.Y() - 0.5,
				playerPos.Z() + (rand.Float32()-0.5)*1.0,
			}
			vel := mgl32.Vec3{0, 2.0, 0}
			color := mgl32.Vec4{0.8, 0.9, 1.0, 0.5} // Light blue
			life := 3.0 + rand.Float32()*2.0
			size := 0.15 + rand.Float32()*0.15
			ps.EmitTyped(bubblePos, vel, color, life, size, ParticleTypeBubble)
		}

		// Fish swimming around
		if len(ps.particles) < 400 && rand.Float32() < 0.05*dt*60 {
			angle := rand.Float32() * 2 * math.Pi
			distance := 5 + rand.Float32()*10
			fishPos := mgl32.Vec3{
				playerPos.X() + float32(math.Cos(float64(angle)))*distance,
				playerPos.Y() + (rand.Float32()-0.5)*3,
				playerPos.Z() + float32(math.Sin(float64(angle)))*distance,
			}
			vel := mgl32.Vec3{
				(rand.Float32() - 0.5) * 3,
				0,
				(rand.Float32() - 0.5) * 3,
			}
			// Different fish colors
			fishColors := []mgl32.Vec4{
				{1.0, 0.5, 0.2, 0.8}, // Orange fish
				{0.3, 0.5, 0.8, 0.8}, // Blue fish
				{0.9, 0.9, 0.3, 0.8}, // Yellow fish
				{0.8, 0.3, 0.3, 0.8}, // Red fish
			}
			color := fishColors[rand.Intn(len(fishColors))]
			life := 10.0 + rand.Float32()*10.0
			size := 0.3 + rand.Float32()*0.2
			ps.EmitTyped(fishPos, vel, color, life, size, ParticleTypeFish)
		}
	} else {
		// Surface bubbles near water
		if rand.Float32() < 0.1*dt*60 {
			bubblePos := mgl32.Vec3{
				playerPos.X() + (rand.Float32()-0.5)*20,
				waterSurfaceY - 0.5,
				playerPos.Z() + (rand.Float32()-0.5)*20,
			}
			vel := mgl32.Vec3{0, 1.5, 0}
			color := mgl32.Vec4{0.9, 0.95, 1.0, 0.4}
			life := 2.0 + rand.Float32()*1.0
			size := 0.1 + rand.Float32()*0.1
			ps.EmitTyped(bubblePos, vel, color, life, size, ParticleTypeBubble)
		}
	}
}
