package render

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// BlockOutlineRenderer handles rendering the selection outline
type BlockOutlineRenderer struct {
	shader *Shader
	vao    uint32
	vbo    uint32
}

// NewBlockOutlineRenderer creates a new block outline renderer
func NewBlockOutlineRenderer() (*BlockOutlineRenderer, error) {
	// Simple shader for the outline
	vertexShader := `
		#version 410 core
		layout (location = 0) in vec3 aPos;
		
		uniform mat4 viewProj;
		uniform vec3 blockPos;
		
		void main() {
			// Add a small offset to avoid z-fighting with the block itself
			vec3 pos = aPos * 1.002 + blockPos - vec3(0.001);
			gl_Position = viewProj * vec4(pos, 1.0);
		}
	`

	fragmentShader := `
		#version 410 core
		out vec4 FragColor;
		
		void main() {
			FragColor = vec4(0.0, 0.0, 0.0, 1.0); // Black outline
		}
	`

	shader, err := NewShader(vertexShader, fragmentShader)
	if err != nil {
		return nil, fmt.Errorf("failed to create outline shader: %w", err)
	}

	renderer := &BlockOutlineRenderer{
		shader: shader,
	}

	renderer.createCubeWireframe()

	return renderer, nil
}

// createCubeWireframe creates the VAO/VBO for a wireframe cube
func (r *BlockOutlineRenderer) createCubeWireframe() {
	// Vertices for a wireframe cube
	vertices := []float32{
		// Bottom face
		0, 0, 0, 1, 0, 0,
		1, 0, 0, 1, 0, 1,
		1, 0, 1, 0, 0, 1,
		0, 0, 1, 0, 0, 0,

		// Top face
		0, 1, 0, 1, 1, 0,
		1, 1, 0, 1, 1, 1,
		1, 1, 1, 0, 1, 1,
		0, 1, 1, 0, 1, 0,

		// Connecting lines
		0, 0, 0, 0, 1, 0,
		1, 0, 0, 1, 1, 0,
		1, 0, 1, 1, 1, 1,
		0, 0, 1, 0, 1, 1,
	}

	gl.GenVertexArrays(1, &r.vao)
	gl.GenBuffers(1, &r.vbo)

	gl.BindVertexArray(r.vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

// Render draws the outline around a block
func (r *BlockOutlineRenderer) Render(blockPos [3]int, viewProj mgl32.Mat4) {
	if r.shader == nil {
		return
	}

	r.shader.Use()
	r.shader.SetMat4("viewProj", viewProj)

	// Set block position uniform
	bPos := mgl32.Vec3{float32(blockPos[0]), float32(blockPos[1]), float32(blockPos[2])}
	r.shader.SetVec3("blockPos", bPos)

	gl.BindVertexArray(r.vao)
	gl.LineWidth(2.0)
	gl.DrawArrays(gl.LINES, 0, 24)
	gl.LineWidth(1.0)
	gl.BindVertexArray(0)
}

// Cleanup releases resources
func (r *BlockOutlineRenderer) Cleanup() {
	if r.shader != nil {
		// r.shader.Delete() // Shader cleanup might be separate
	}
	gl.DeleteVertexArrays(1, &r.vao)
	gl.DeleteBuffers(1, &r.vbo)
}
