package render

import (
	"fmt"
	"math"
	"math/rand"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
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
	// 1. Create Shader
	vShaderFile := "assets/shaders/particle.vert"
	fShaderFile := "assets/shaders/particle.frag"

	// Read shader files (assuming they exist or we create them)
	// For now, let's defer reading if they are not created yet.
	// But we should try.
	vSource, err := os.ReadFile(vShaderFile)
	if err == nil {
		fSource, _ := os.ReadFile(fShaderFile)
		ps.shader, err = NewShader(string(vSource), string(fSource))
		if err != nil {
			fmt.Printf("Warning: Failed to compile particle shader: %v\n", err)
		}
	} else {
		// Create default if file missing (dev mode)
		// ... (省略 for brevity, assumes files will be created)
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
