package render

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// BlockBreaker handles the block breaking animation
type BlockBreaker struct {
	shader        *Shader
	vao           uint32
	vbo           uint32
	progress      float32
	isBreaking    bool
	breakingBlock [3]int
	breakTime     float32
	currentSpeed  float32 // 1.0 / breakTime
}

// NewBlockBreaker creates a new block breaker
func NewBlockBreaker() (*BlockBreaker, error) {
	// Shader for the breaking overlay
	// We'll use a procedural noise/pattern for cracks to avoid needing textures for now
	vertexShader := `
		#version 410 core
		layout (location = 0) in vec3 aPos;
		layout (location = 1) in vec2 aTexCoord;
		
		uniform mat4 viewProj;
		uniform vec3 blockPos;
		
		out vec2 TexCoord;
		out vec3 LocalPos;
		
		void main() {
			// Add a very small offset to avoid z-fighting with the block
			vec3 pos = aPos * 1.001 + blockPos - vec3(0.0005);
			gl_Position = viewProj * vec4(pos, 1.0);
			TexCoord = aTexCoord;
			LocalPos = aPos;
		}
	`

	fragmentShader := `
		#version 410 core
		out vec4 FragColor;
		
		in vec2 TexCoord;
		in vec3 LocalPos;
		
		uniform float progress; // 0.0 to 1.0
		
		// Simple hash function for randomness
		float hash(vec2 col) {
			return fract(sin(dot(col, vec2(12.9898, 78.233))) * 43758.5453);
		}
		
		void main() {
			// Procedural crack generation based on progress
			// We define 10 stages of cracks
			
			if (progress <= 0.0) discard;
			
			// Center UVs
			vec2 uv = TexCoord * 2.0 - 1.0;
			float dist = length(uv);
			
			// Noise for jaggy lines
			float noise = hash(floor(uv * 10.0));
			
			// Check if this pixel is part of a crack
			// This is a very simplified procedural crack pattern
			// In a real implementation, we would use textures
			
			// Create a web-like pattern from center
			float angle = atan(uv.y, uv.x);
			float crackId = floor(angle * 3.0 + noise);
			
			float stage = floor(progress * 10.0) / 10.0;
			
			// Dark color for cracks
			vec4 crackColor = vec4(0.0, 0.0, 0.0, 0.7);
			
			// Main cracks radiating from center
			if (abs(sin(angle * 5.0 + noise)) < 0.1 * stage) {
				FragColor = crackColor;
				return;
			}
			
			// Concentric cracks
			if (abs(dist - stage * 0.8) < 0.05) {
				FragColor = crackColor;
				return;
			}
			
			// Random noise cracks
			if (hash(uv * 20.0) < stage * 0.2) {
				FragColor = crackColor;
				return;
			}
			
			discard;
		}
	`

	shader, err := NewShader(vertexShader, fragmentShader)
	if err != nil {
		return nil, fmt.Errorf("failed to create breaking shader: %w", err)
	}

	breaker := &BlockBreaker{
		shader: shader,
	}

	breaker.createCubeMesh()

	return breaker, nil
}

// createCubeMesh creates the VAO/VBO for a cube mesh with UVs
func (b *BlockBreaker) createCubeMesh() {
	// Vertices for a standard cube with UVs
	// Position (3) + TexCoord (2)
	vertices := []float32{
		// Back face
		0, 0, 0, 0, 0,
		1, 1, 0, 1, 1,
		1, 0, 0, 1, 0,
		0, 0, 0, 0, 0,
		0, 1, 0, 0, 1,
		1, 1, 0, 1, 1,

		// Front face
		0, 0, 1, 0, 0,
		1, 0, 1, 1, 0,
		1, 1, 1, 1, 1,
		0, 0, 1, 0, 0,
		1, 1, 1, 1, 1,
		0, 1, 1, 0, 1,

		// Left face
		0, 1, 1, 1, 0,
		0, 1, 0, 1, 1,
		0, 0, 0, 0, 1,
		0, 1, 1, 1, 0,
		0, 0, 0, 0, 1,
		0, 0, 1, 0, 0,

		// Right face
		1, 1, 1, 1, 0,
		1, 0, 0, 0, 1,
		1, 1, 0, 1, 1,
		1, 1, 1, 1, 0,
		1, 0, 1, 0, 0,
		1, 0, 0, 0, 1,

		// Bottom face
		0, 0, 0, 0, 1,
		1, 0, 0, 1, 1,
		1, 0, 1, 1, 0,
		0, 0, 0, 0, 1,
		1, 0, 1, 1, 0,
		0, 0, 1, 0, 0,

		// Top face
		0, 1, 0, 0, 1,
		0, 1, 1, 0, 0,
		1, 1, 1, 1, 0,
		0, 1, 0, 0, 1,
		1, 1, 1, 1, 0,
		1, 1, 0, 1, 1,
	}

	gl.GenVertexArrays(1, &b.vao)
	gl.GenBuffers(1, &b.vbo)

	gl.BindVertexArray(b.vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Position attribute
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// TexCoord attribute
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

// StartBreaking starts the breaking process for a block
func (b *BlockBreaker) StartBreaking(blockPos [3]int, breakTime float32) {
	if b.breakingBlock != blockPos {
		b.breakingBlock = blockPos
		b.progress = 0
	} else if !b.isBreaking {
		// Resume or reset? Let's check logic. Usually, if you stop breaking, it resets.
		// So if we are starting again, we might want to ensure we are in breaking state.
		// If it's a new block (handled above), reset.
		// If same block but wasn't breaking, maybe we keep progress or reset?
		// Minecraft resets if you stop.
	}

	b.isBreaking = true
	b.breakTime = breakTime
	if breakTime <= 0 {
		b.currentSpeed = 1000 // Instant
	} else {
		b.currentSpeed = 1.0 / breakTime
	}
}

// StopBreaking stops the breaking process and resets progress
func (b *BlockBreaker) StopBreaking() {
	b.isBreaking = false
	b.progress = 0
}

// Update updates the breaking progress
// Returns true if the block is broken
func (b *BlockBreaker) Update(dt float32) bool {
	if !b.isBreaking {
		b.progress = 0
		return false
	}

	b.progress += b.currentSpeed * dt

	if b.progress >= 1.0 {
		b.progress = 0
		return true // Block broken
	}

	return false
}

// Render draws the breaking overlay
func (b *BlockBreaker) Render(viewProj mgl32.Mat4) {
	if !b.isBreaking || b.progress <= 0 || b.shader == nil {
		return
	}

	b.shader.Use()
	b.shader.SetMat4("viewProj", viewProj)

	// Set block position uniform
	bPos := mgl32.Vec3{float32(b.breakingBlock[0]), float32(b.breakingBlock[1]), float32(b.breakingBlock[2])}
	b.shader.SetVec3("blockPos", bPos)

	b.shader.SetFloat("progress", b.progress)

	gl.BindVertexArray(b.vao)

	// Enable blending for transparency
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.DrawArrays(gl.TRIANGLES, 0, 36)

	gl.Disable(gl.BLEND)
	gl.BindVertexArray(0)
}

// Cleanup releases resources
func (b *BlockBreaker) Cleanup() {
	if b.shader != nil {
		// b.shader.Delete()
	}
	gl.DeleteVertexArrays(1, &b.vao)
	gl.DeleteBuffers(1, &b.vbo)
}
