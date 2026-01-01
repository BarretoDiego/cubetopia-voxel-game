// Package ui provides user interface rendering
package ui

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Renderer handles UI rendering with OpenGL
type Renderer struct {
	// Screen dimensions
	width  int
	height int

	// Shader for UI rendering
	shader *UIShader

	// VAO/VBO for quads
	quadVAO uint32
	quadVBO uint32

	// Font rendering (simplified - just colored rectangles for now)
}

// UIShader is a simple shader for UI elements
type UIShader struct {
	ID uint32
}

// NewRenderer creates a new UI renderer
func NewRenderer(width, height int) (*Renderer, error) {
	r := &Renderer{
		width:  width,
		height: height,
	}

	// Create shader
	shader, err := createUIShader()
	if err != nil {
		return nil, err
	}
	r.shader = shader

	// Create quad mesh
	r.createQuadMesh()

	return r, nil
}

// Resize updates the screen dimensions
func (r *Renderer) Resize(width, height int) {
	r.width = width
	r.height = height
}

// BeginFrame prepares for UI rendering
func (r *Renderer) BeginFrame() {
	// Switch to 2D orthographic mode
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

// EndFrame finishes UI rendering
func (r *Renderer) EndFrame() {
	gl.Enable(gl.DEPTH_TEST)
}

// DrawRect draws a colored rectangle
func (r *Renderer) DrawRect(x, y, width, height float32, color [4]float32) {
	if r.shader == nil {
		return
	}

	gl.UseProgram(r.shader.ID)

	// Set uniforms
	projection := mgl32.Ortho(0, float32(r.width), float32(r.height), 0, -1, 1)
	projLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uProjection\x00"))
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

	model := mgl32.Translate3D(x, y, 0).Mul4(mgl32.Scale3D(width, height, 1))
	modelLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uModel\x00"))
	gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])

	colorLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uColor\x00"))
	gl.Uniform4fv(colorLoc, 1, &color[0])

	gl.BindVertexArray(r.quadVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}

// DrawCrosshair draws a crosshair at the center of the screen
func (r *Renderer) DrawCrosshair() {
	cx := float32(r.width) / 2
	cy := float32(r.height) / 2
	size := float32(10)
	thickness := float32(2)

	white := [4]float32{1, 1, 1, 0.8}

	// Horizontal line
	r.DrawRect(cx-size, cy-thickness/2, size*2, thickness, white)
	// Vertical line
	r.DrawRect(cx-thickness/2, cy-size, thickness, size*2, white)
}

// DrawHotbar draws the hotbar at the bottom of the screen
func (r *Renderer) DrawHotbar(selectedIndex int, blockColors [][3]float32) {
	slotSize := float32(50)
	padding := float32(4)
	numSlots := len(blockColors)
	if numSlots > 9 {
		numSlots = 9
	}

	totalWidth := float32(numSlots) * (slotSize + padding)
	startX := (float32(r.width) - totalWidth) / 2
	startY := float32(r.height) - slotSize - 20

	for i := 0; i < numSlots; i++ {
		x := startX + float32(i)*(slotSize+padding)

		// Slot background
		bgColor := [4]float32{0, 0, 0, 0.5}
		if i == selectedIndex {
			bgColor = [4]float32{1, 1, 1, 0.3}
		}
		r.DrawRect(x, startY, slotSize, slotSize, bgColor)

		// Block color preview
		if i < len(blockColors) {
			color := blockColors[i]
			blockColor := [4]float32{color[0], color[1], color[2], 1.0}
			r.DrawRect(x+8, startY+8, slotSize-16, slotSize-16, blockColor)
		}

		// Selection border
		if i == selectedIndex {
			r.DrawRect(x, startY, slotSize, 2, [4]float32{1, 1, 0, 1})            // Top
			r.DrawRect(x, startY+slotSize-2, slotSize, 2, [4]float32{1, 1, 0, 1}) // Bottom
			r.DrawRect(x, startY, 2, slotSize, [4]float32{1, 1, 0, 1})            // Left
			r.DrawRect(x+slotSize-2, startY, 2, slotSize, [4]float32{1, 1, 0, 1}) // Right
		}
	}
}

// DebugInfo contains debug information to display
type DebugInfo struct {
	Position     mgl32.Vec3
	ChunksLoaded int
	FPS          int
	Biome        string
	MemoryMB     int
}

// DrawDebugPanel draws debug information
func (r *Renderer) DrawDebugPanel(info DebugInfo) {
	x := float32(10)
	y := float32(10)
	width := float32(200)
	lineHeight := float32(20)
	padding := float32(10)

	// Background
	lines := 6
	height := float32(lines)*lineHeight + padding*2
	r.DrawRect(x, y, width, height, [4]float32{0, 0, 0, 0.6})

	// We can't easily draw text without a font system
	// For now, we'll just show colored indicators

	// FPS indicator (green = good, yellow = ok, red = bad)
	fpsColor := [4]float32{0, 1, 0, 1} // Green
	if info.FPS < 30 {
		fpsColor = [4]float32{1, 0, 0, 1} // Red
	} else if info.FPS < 60 {
		fpsColor = [4]float32{1, 1, 0, 1} // Yellow
	}
	r.DrawRect(x+padding, y+padding, float32(min(info.FPS, 120)), 10, fpsColor)

	// Chunks indicator
	chunkColor := [4]float32{0.4, 0.6, 1, 1}
	r.DrawRect(x+padding, y+padding+lineHeight, float32(info.ChunksLoaded*2), 10, chunkColor)
}

// Cleanup releases resources
func (r *Renderer) Cleanup() {
	if r.quadVAO != 0 {
		gl.DeleteVertexArrays(1, &r.quadVAO)
	}
	if r.quadVBO != 0 {
		gl.DeleteBuffers(1, &r.quadVBO)
	}
	if r.shader != nil && r.shader.ID != 0 {
		gl.DeleteProgram(r.shader.ID)
	}
}

func (r *Renderer) createQuadMesh() {
	// Unit quad vertices (2D positions)
	vertices := []float32{
		0, 0,
		1, 0,
		1, 1,
		0, 0,
		1, 1,
		0, 1,
	}

	gl.GenVertexArrays(1, &r.quadVAO)
	gl.GenBuffers(1, &r.quadVBO)

	gl.BindVertexArray(r.quadVAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, r.quadVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointerWithOffset(0, 2, gl.FLOAT, false, 2*4, 0)
	gl.EnableVertexAttribArray(0)

	gl.BindVertexArray(0)
}

func createUIShader() (*UIShader, error) {
	vertexSource := `
#version 410 core
layout(location = 0) in vec2 aPos;

uniform mat4 uProjection;
uniform mat4 uModel;

void main() {
    gl_Position = uProjection * uModel * vec4(aPos, 0.0, 1.0);
}
` + "\x00"

	fragmentSource := `
#version 410 core
uniform vec4 uColor;
out vec4 fragColor;

void main() {
    fragColor = uColor;
}
` + "\x00"

	// Compile shaders
	vertexShader, err := compileShader(vertexSource, gl.VERTEX_SHADER)
	if err != nil {
		return nil, fmt.Errorf("vertex shader: %w", err)
	}
	defer gl.DeleteShader(vertexShader)

	fragmentShader, err := compileShader(fragmentSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, fmt.Errorf("fragment shader: %w", err)
	}
	defer gl.DeleteShader(fragmentShader)

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		return nil, fmt.Errorf("failed to link UI shader")
	}

	return &UIShader{ID: program}, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		return 0, fmt.Errorf("failed to compile shader")
	}

	return shader, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
