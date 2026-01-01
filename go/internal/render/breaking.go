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
	vertexShader := `
		#version 410 core
		layout (location = 0) in vec3 aPos;
		layout (location = 1) in vec2 aTexCoord;
		
		uniform mat4 viewProj;
		uniform vec3 blockPos;
		uniform float progress;
		uniform float uTime;
		
		out vec2 TexCoord;
		
		void main() {
			// Shaking effect
			float shake = sin(uTime * 40.0) * progress * 0.02;
			vec3 offset = vec3(shake, shake * 0.5, shake * 0.3);
			
			// Slightly larger than a block to avoid z-fighting
			vec3 pos = aPos * 1.002 + blockPos - vec3(0.001) + offset;
			gl_Position = viewProj * vec4(pos, 1.0);
			TexCoord = aTexCoord;
		}
	`

	fragmentShader := `
		#version 410 core
		out vec4 FragColor;
		
		in vec2 TexCoord;
		
		uniform float progress; // 0.0 to 1.0
		
		float hash(vec2 p) {
			return fract(sin(dot(p, vec2(12.9898, 78.233))) * 43758.5453);
		}
		
		float noise(vec2 p) {
			vec2 i = floor(p);
			vec2 f = fract(p);
			f = f*f*(3.0-2.0*f);
			float a = hash(i);
			float b = hash(i + vec2(1.0, 0.0));
			float c = hash(i + vec2(0.0, 1.0));
			float d = hash(i + vec2(1.0, 1.0));
			return mix(mix(a, b, f.x), mix(c, d, f.x), f.y);
		}
		
		void main() {
			if (progress <= 0.0) discard;
			
			vec2 uv = TexCoord;
			
			// Darkening based on progress
			float darken = progress * 0.6;
			
			// Procedural cracks
			float n = noise(uv * 12.0);
			float n2 = noise(uv * 24.0 + n);
			
			// Distance from center for radial cracks
			vec2 distVec = uv - vec2(0.5);
			float dist = length(distVec);
			float angle = atan(distVec.y, distVec.x);
			
			// Main jagged lines
			float crack = abs(n - 0.5) * 2.0;
			crack *= abs(n2 - 0.5) * 4.0;
			
			// Radial cracks from center
			float radial = abs(sin(angle * 4.0 + n * 2.0));
			radial = pow(radial, 10.0) * (progress * 2.1);
			
			// Crack width based on progress
			float threshold = 1.0 - (progress * 0.85);
			
			bool isCrack = (crack > threshold) || (radial > 1.2 - progress);
			
			if (isCrack) {
				FragColor = vec4(0.0, 0.0, 0.0, 0.85);
			} else {
				// Base face darkening
				FragColor = vec4(0.0, 0.0, 0.0, darken);
			}
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
		b.progress = 1.0 // Cap it
		return true      // Block broken
	}

	return false
}

// IsBreaking returns true if the breaker is active
func (b *BlockBreaker) IsBreaking() bool {
	return b.isBreaking
}

// GetBreakingBlock returns the position of the block being broken
func (b *BlockBreaker) GetBreakingBlock() [3]int {
	return b.breakingBlock
}

// Render draws the breaking overlay
func (b *BlockBreaker) Render(viewProj mgl32.Mat4, time float32) {
	if !b.isBreaking || b.progress <= 0 || b.shader == nil {
		return
	}

	b.shader.Use()
	b.shader.SetMat4("viewProj", viewProj)

	// Set block position uniform
	bPos := mgl32.Vec3{float32(b.breakingBlock[0]), float32(b.breakingBlock[1]), float32(b.breakingBlock[2])}
	b.shader.SetVec3("blockPos", bPos)

	b.shader.SetFloat("progress", b.progress)
	b.shader.SetFloat("uTime", time)

	gl.BindVertexArray(b.vao)

	// Enable blending for transparency
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Disable depth writing but keep depth testing
	gl.DepthMask(false)

	gl.DrawArrays(gl.TRIANGLES, 0, 36)

	gl.DepthMask(true)
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
