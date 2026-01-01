// Package ui provides user interface rendering
package ui

import (
	"fmt"
	"voxelgame/internal/core/block"

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

	// VAO/VBO for 3D items in UI
	cubeVAO uint32
	cubeVBO uint32

	// Font rendering
	font *Font
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

	// Create cube mesh for 3D items
	r.createCubeMesh()

	// Initialize font
	r.font = NewFont()

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

	useTexLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uUseTexture\x00"))
	gl.Uniform1i(useTexLoc, 0) // No texture

	gl.BindVertexArray(r.quadVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}

// DrawText draws text string at position
func (r *Renderer) DrawText(x, y, scale float32, text string, color [4]float32) {
	if r.font == nil || r.shader == nil {
		return
	}

	gl.UseProgram(r.shader.ID)
	r.font.Bind()

	// Set uniforms common to all chars
	projection := mgl32.Ortho(0, float32(r.width), float32(r.height), 0, -1, 1)
	projLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uProjection\x00"))
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

	colorLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uColor\x00"))
	gl.Uniform4fv(colorLoc, 1, &color[0])

	useTexLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uUseTexture\x00"))
	gl.Uniform1i(useTexLoc, 1) // Use texture

	texLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uTexture\x00"))
	gl.Uniform1i(texLoc, 0) // Texture unit 0

	// Draw each character
	cursorX := x
	charWidth := float32(r.font.charWidth) * scale
	charHeight := float32(r.font.charHeight) * scale

	// Uniform locations
	modelLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uModel\x00"))
	uvLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uUVBounds\x00"))

	gl.BindVertexArray(r.quadVAO)

	for _, char := range text {
		u1, v1, u2, v2 := r.font.CharUV(char)

		// Set UVs for this character via uniform (since quad mesh is static 0..1)
		// We modify the shader to scale UVs? No, easier to pass UV range uniform.
		gl.Uniform4f(uvLoc, u1, v1, u2, v2)

		model := mgl32.Translate3D(cursorX, y, 0).Mul4(mgl32.Scale3D(charWidth, charHeight, 1))
		gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])

		gl.DrawArrays(gl.TRIANGLES, 0, 6)
		cursorX += (charWidth + 1*scale)
	}
	gl.BindVertexArray(0)
}

// DrawCrosshair draws a crosshair at the center of the screen
func (r *Renderer) DrawCrosshair() {
	centerX := float32(r.width) / 2
	centerY := float32(r.height) / 2
	size := float32(10.0)
	thickness := float32(2.0)
	color := [4]float32{1, 1, 1, 0.8}

	// Horizontal line
	r.DrawRect(centerX-size, centerY-thickness/2, size*2, thickness, color)
	// Vertical line
	r.DrawRect(centerX-thickness/2, centerY-size, thickness, size*2, color)
}

// DrawTargetInfo displays the name of the targeted block/entity
func (r *Renderer) DrawTargetInfo(name string) {
	if name == "" {
		return
	}

	centerX := float32(r.width) / 2
	centerY := float32(r.height) / 2

	// Position text below crosshair
	textY := centerY + 30.0
	scale := float32(1.0) // Slightly smaller than default

	// Estimate text width to center it (approximate)
	// Font rendering doesn't give us easy width measurement yet, so we'll guess based on char count
	estimatedWidth := float32(len(name)) * 10.0 // Approx 10px per char
	textX := centerX - estimatedWidth/2

	// Draw background for readability
	// Padding
	padding := float32(8.0)
	bgWidth := estimatedWidth + padding*2
	bgHeight := float32(24.0)
	bgColor := [4]float32{0, 0, 0, 0.5}

	r.DrawRect(centerX-bgWidth/2, textY-4, bgWidth, bgHeight, bgColor)

	// Draw text
	r.DrawText(textX, textY, scale, name, [4]float32{1, 1, 1, 1})
}

// DrawHotbar draws the hotbar at the bottom of the screen
func (r *Renderer) DrawHotbar(selectedIndex int, blockColors [][3]float32, blockTypes []block.Type) {
	r.DrawHotbarWithCounts(selectedIndex, blockColors, blockTypes, nil, nil)
}

// DrawHotbarWithCounts draws the hotbar with item quantities and selected name
func (r *Renderer) DrawHotbarWithCounts(selectedIndex int, blockColors [][3]float32, blockTypes []block.Type, counts []int, names []string) {
	slotSize := float32(60) // Increased from 50
	padding := float32(6)
	numSlots := 9 // Fixed at 9

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

		// Block color preview (3D)
		if i < len(blockColors) {
			color := blockColors[i]
			bType := block.Air
			if blockTypes != nil && i < len(blockTypes) {
				bType = blockTypes[i]
			} else {
				// Fallback to finding type by color is hard, assume standard block if we don't have types?
				// Actually the caller should pass types. For now default to standard cube render if generic.
			}

			// Render 3D item
			r.Render3DItemInBox(x, startY, slotSize, bType, color)
		}

		// Selection border
		if i == selectedIndex {
			r.DrawRect(x, startY, slotSize, 2, [4]float32{1, 1, 0, 1})            // Top
			r.DrawRect(x, startY+slotSize-2, slotSize, 2, [4]float32{1, 1, 0, 1}) // Bottom
			r.DrawRect(x, startY, 2, slotSize, [4]float32{1, 1, 0, 1})            // Left
			r.DrawRect(x+slotSize-2, startY, 2, slotSize, [4]float32{1, 1, 0, 1}) // Right
		}

		// Slot Number (top-left)
		numStr := fmt.Sprintf("%d", i+1)
		r.DrawText(x+2, startY+2, 1.0, numStr, [4]float32{1, 1, 1, 0.8})

		// Item Count (bottom-right)
		if counts != nil && i < len(counts) && counts[i] > 0 {
			countStr := fmt.Sprintf("%d", counts[i])
			// Position in bottom-right corner
			r.DrawText(x+slotSize-float32(len(countStr)*8)-4, startY+slotSize-14, 1.0, countStr, [4]float32{1, 1, 1, 1})
		}
	}

	// Draw selected item name above hotbar
	if names != nil && selectedIndex >= 0 && selectedIndex < len(names) {
		name := names[selectedIndex]
		if name != "" {
			nameScale := float32(1.5)
			estimatedWidth := float32(len(name)) * 12.0 * nameScale / 1.5 // Approx
			nameX := (float32(r.width) - estimatedWidth) / 2
			nameY := startY - 30

			// Subtle shadow/background for text
			r.DrawRect(nameX-4, nameY-2, estimatedWidth+8, 22, [4]float32{0, 0, 0, 0.4})
			r.DrawText(nameX, nameY, nameScale, name, [4]float32{1, 1, 0.8, 1})
		}
	}
}

// BlockDisplayInfo contains info for displaying a block in the inventory panel
type BlockDisplayInfo struct {
	Type       block.Type
	Color      [3]float32
	Name       string
	Count      int
	HotbarSlot int // -1 if not in hotbar
}

// DrawInventoryPanel draws the expanded inventory with all blocks
func (r *Renderer) DrawInventoryPanel(blocks []BlockDisplayInfo, selectedHotbarIndex int, panelSelectedIndex int) {
	if r.shader == nil || len(blocks) == 0 {
		return
	}

	// Panel dimensions - Increased for better visibility
	cols := 8 // Increased from 6
	rows := (len(blocks) + cols - 1) / cols
	slotSize := float32(70) // Increased from 60
	padding := float32(8)
	panelPadding := float32(30)

	panelWidth := float32(cols)*(slotSize+padding) + panelPadding*2
	panelHeight := float32(rows)*(slotSize+padding) + panelPadding*2 + 40 // Extra for title

	// Center the panel
	panelX := (float32(r.width) - panelWidth) / 2
	panelY := (float32(r.height) - panelHeight) / 2

	// Background
	r.DrawRect(panelX, panelY, panelWidth, panelHeight, [4]float32{0.08, 0.08, 0.1, 0.95})

	// Border
	borderColor := [4]float32{0.3, 0.6, 0.4, 1}
	r.DrawRect(panelX, panelY, panelWidth, 3, borderColor)               // Top
	r.DrawRect(panelX, panelY+panelHeight-3, panelWidth, 3, borderColor) // Bottom
	r.DrawRect(panelX, panelY, 3, panelHeight, borderColor)              // Left
	r.DrawRect(panelX+panelWidth-3, panelY, 3, panelHeight, borderColor) // Right

	// Title
	r.DrawText(panelX+panelPadding, panelY+10, 1.8, "INVENTARIO (I para fechar)", [4]float32{1, 1, 0.5, 1})

	// Draw blocks in grid
	startX := panelX + panelPadding
	startY := panelY + 50

	for i, block := range blocks {
		col := i % cols
		row := i / cols

		x := startX + float32(col)*(slotSize+padding)
		y := startY + float32(row)*(slotSize+padding)

		// Slot background
		bgColor := [4]float32{0.15, 0.15, 0.2, 0.8}
		if i == panelSelectedIndex {
			// Highlight current selection
			bgColor = [4]float32{0.3, 0.4, 0.6, 1.0}
		} else if block.HotbarSlot >= 0 {
			// Highlight blocks in hotbar
			bgColor = [4]float32{0.2, 0.3, 0.25, 0.9}
		}
		r.DrawRect(x, y, slotSize, slotSize, bgColor)

		// Border for selection
		if i == panelSelectedIndex {
			r.DrawRect(x, y, slotSize, 2, [4]float32{1, 1, 0, 1})            // Top
			r.DrawRect(x, y+slotSize-2, slotSize, 2, [4]float32{1, 1, 0, 1}) // Bottom
			r.DrawRect(x, y, 2, slotSize, [4]float32{1, 1, 0, 1})            // Left
			r.DrawRect(x+slotSize-2, y, 2, slotSize, [4]float32{1, 1, 0, 1}) // Right
		}

		// Draw block preview (3D)
		r.Render3DItemInBox(x, y, slotSize, block.Type, block.Color)

		// Hotbar slot indicator (top-left)
		if block.HotbarSlot >= 0 {
			hotbarStr := fmt.Sprintf("[%d]", block.HotbarSlot+1)
			r.DrawText(x+4, y+4, 1.0, hotbarStr, [4]float32{1, 1, 0, 1})
		}

		// Count (bottom-right)
		if block.Count > 0 {
			countStr := fmt.Sprintf("%d", block.Count)
			r.DrawText(x+slotSize-float32(len(countStr)*8)-4, y+slotSize-16, 1.0, countStr, [4]float32{1, 1, 1, 1})
		} else {
			// Show "0" for items not in inventory
			r.DrawText(x+slotSize-12, y+slotSize-16, 0.9, "0", [4]float32{0.5, 0.5, 0.5, 0.8})
		}
	}

	// Name of the highlighted item
	if panelSelectedIndex >= 0 && panelSelectedIndex < len(blocks) {
		selectedItem := blocks[panelSelectedIndex]
		nameText := "Item: " + selectedItem.Name
		r.DrawText(panelX+panelPadding, startY+float32(rows)*(slotSize+padding)+10, 1.4, nameText, [4]float32{1, 1, 0, 1})
	}

	// Instructions at bottom
	r.DrawText(panelX+panelPadding, panelY+panelHeight-30, 1.2, "Teclas 1-9: Selecionar item da hotbar", [4]float32{0.7, 0.7, 0.7, 1})
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

	// Text Info
	white := [4]float32{1, 1, 1, 1}

	// FPS
	r.DrawText(x+padding, y+padding, 1.5, fmt.Sprintf("FPS: %d", info.FPS), white)

	// Chunks
	r.DrawText(x+padding, y+padding+lineHeight, 1.5, fmt.Sprintf("Chunks: %d", info.ChunksLoaded), white)

	// Position
	posStr := fmt.Sprintf("Pos: %.1f, %.1f, %.1f", info.Position.X(), info.Position.Y(), info.Position.Z())
	r.DrawText(x+padding, y+padding+lineHeight*2, 1.5, posStr, white)

	// Biome
	r.DrawText(x+padding, y+padding+lineHeight*3, 1.5, fmt.Sprintf("Biome: %s", info.Biome), white)

	// Memory
	r.DrawText(x+padding, y+padding+lineHeight*4, 1.5, fmt.Sprintf("Mem: %d MB", info.MemoryMB), white)
}

// DrawControlsOverlay draws a list of game controls
func (r *Renderer) DrawControlsOverlay(commands []string) {
	if r.shader == nil || len(commands) == 0 {
		return
	}

	width := float32(300)
	lineHeight := float32(20)
	padding := float32(10)

	height := float32(len(commands))*lineHeight + padding*2

	// Position: Top Right, below minimap (approx 150px down)
	x := float32(r.width) - width - 10
	y := float32(160)

	// Background
	r.DrawRect(x, y, width, height, [4]float32{0, 0, 0, 0.7})

	// Header
	r.DrawText(x+padding, y+padding, 1.2, "CONTROLS (H to toggle)", [4]float32{1, 1, 0, 1})

	// Commands
	white := [4]float32{1, 1, 1, 1}
	for i, cmd := range commands {
		r.DrawText(x+padding, y+padding+float32(i+1)*lineHeight, 1.2, cmd, white)
	}
}

// DrawTimeIndicator renders the current time
func (r *Renderer) DrawTimeIndicator(timeString string) {
	if r.shader == nil {
		return
	}

	width := float32(120)
	height := float32(36)

	// Position: Top Center
	x := (float32(r.width) - width) / 2
	y := float32(10)

	// Background
	r.DrawRect(x, y, width, height, [4]float32{0, 0, 0, 0.5})

	// Border
	r.DrawRect(x, y, width, 2, [4]float32{0.5, 0.5, 0.5, 0.8})          // Top
	r.DrawRect(x, y+height-2, width, 2, [4]float32{0.5, 0.5, 0.5, 0.8}) // Bottom
	r.DrawRect(x, y, 2, height, [4]float32{0.5, 0.5, 0.5, 0.8})         // Left
	r.DrawRect(x+width-2, y, 2, height, [4]float32{0.5, 0.5, 0.5, 0.8}) // Right

	// Icon (Sun/Moon) or just text
	// Center text
	textWidth := float32(len(timeString) * 12) // Approx
	textX := x + (width-textWidth)/2

	r.DrawText(textX, y+8, 1.2, timeString, [4]float32{1, 0.9, 0.6, 1})
}

// DrawIsometricCube renders a fake 3D cube for UI
func (r *Renderer) DrawIsometricCube(x, y, size float32, color [3]float32) {
	if r.shader == nil {
		return
	}

	half := size / 2
	quarter := size / 4

	// Colors
	topColor := [4]float32{color[0] * 1.2, color[1] * 1.2, color[2] * 1.2, 1.0}
	frontColor := [4]float32{color[0], color[1], color[2], 1.0}
	sideColor := [4]float32{color[0] * 0.8, color[1] * 0.8, color[2] * 0.8, 1.0} // Darker

	// Top Face (Rhombus)
	// Center is x+half, y+quarter
	// P1: Top (x+half, y)
	// P2: Right (x+size, y+quarter)
	// P3: Bottom (x+half, y+half)
	// P4: Left (x, y+quarter)
	r.drawQuadCustom(
		x+half, y, // Top
		x+size, y+quarter, // Right
		x+half, y+half, // Bottom
		x, y+quarter, // Left
		topColor,
	)

	// Front Face (Right)
	// P1: Top-Left (x+half, y+half)
	// P2: Top-Right (x+size, y+quarter)
	// P3: Bot-Right (x+size, y+size-quarter)
	// P4: Bot-Left (x+half, y+size)
	r.drawQuadCustom(
		x+half, y+half,
		x+size, y+quarter,
		x+size, y+size-quarter,
		x+half, y+size,
		sideColor,
	)

	// Side Face (Left)
	r.drawQuadCustom(
		x, y+quarter,
		x+half, y+half,
		x+half, y+size,
		x, y+size-quarter,
		frontColor,
	)
}

// drawQuadCustom draws a generic quad given 4 points
func (r *Renderer) drawQuadCustom(x1, y1, x2, y2, x3, y3, x4, y4 float32, color [4]float32) {
	gl.UseProgram(r.shader.ID)

	// Set uniforms
	// Identity Model
	model := mgl32.Ident4()
	modLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uModel\x00"))
	gl.UniformMatrix4fv(modLoc, 1, false, &model[0])

	projection := mgl32.Ortho(0, float32(r.width), float32(r.height), 0, -1, 1)
	projLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uProjection\x00"))
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

	colorLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uColor\x00"))
	gl.Uniform4fv(colorLoc, 1, &color[0])

	useTexLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uUseTexture\x00"))
	gl.Uniform1i(useTexLoc, 0)

	// Dynamic vertices
	vertices := []float32{
		x1, y1, 0, 0,
		x2, y2, 0, 0,
		x3, y3, 0, 0,
		x1, y1, 0, 0,
		x3, y3, 0, 0,
		x4, y4, 0, 0,
	}

	// Update VBO - Streaming
	gl.BindBuffer(gl.ARRAY_BUFFER, r.quadVBO)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(vertices)*4, gl.Ptr(vertices))

	gl.BindVertexArray(r.quadVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)

	// Restore VBO to unit quad?
	// Actually, DrawRect relies on 0..1 unit quad and uses transforms.
	// If I overwrite VBO, I break DrawRect!
	// Solution: Create a separate `dynamicVAO` or just use a new VBO for custom.
	// OR: Restore the unit quad data after use.
	// OR: Use immediate mode (deprecated).
	// Safer: Restore unit quad.
	r.restoreQuadVBO()
}

func (r *Renderer) restoreQuadVBO() {
	vertices := []float32{
		0, 0, 0, 0,
		1, 0, 1, 0,
		1, 1, 1, 1,
		0, 0, 0, 0,
		1, 1, 1, 1,
		0, 1, 0, 1,
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, r.quadVBO)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(vertices)*4, gl.Ptr(vertices))
}

// DrawMinimap draws the minimap as a radar in the bottom-left corner
func (r *Renderer) DrawMinimap(textureID uint32) {
	if r.shader == nil {
		return
	}

	size := float32(256) // Larger radar
	margin := float32(16)
	borderWidth := float32(6)

	// Position: Bottom-Left
	x := margin
	y := float32(r.height) - size - margin - 80 // Account for hotbar

	gl.UseProgram(r.shader.ID)

	// Outer border (dark frame)
	r.DrawRect(x-borderWidth-2, y-borderWidth-2, size+borderWidth*2+4, size+borderWidth*2+4, [4]float32{0.05, 0.1, 0.08, 1.0})

	// Border (metallic green)
	r.DrawRect(x-borderWidth, y-borderWidth, size+borderWidth*2, size+borderWidth*2, [4]float32{0.15, 0.25, 0.18, 1.0})

	// Inner border highlight
	r.DrawRect(x-2, y-2, size+4, size+4, [4]float32{0.0, 0.4, 0.25, 0.8})

	// Bind Texture
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, textureID)

	// Projection
	projection := mgl32.Ortho(0, float32(r.width), float32(r.height), 0, -1, 1)
	projLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uProjection\x00"))
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

	// Color (Tint)
	colLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uColor\x00"))
	gl.Uniform4f(colLoc, 1, 1, 1, 1)

	// Texture flags
	useTexLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uUseTexture\x00"))
	gl.Uniform1i(useTexLoc, 1)

	texLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uTexture\x00"))
	gl.Uniform1i(texLoc, 0)

	uvLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uUVBounds\x00"))
	gl.Uniform4f(uvLoc, 0, 0, 1, 1)

	// Model
	model := mgl32.Translate3D(x, y, 0).Mul4(mgl32.Scale3D(size, size, 1))
	modLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uModel\x00"))
	gl.UniformMatrix4fv(modLoc, 1, false, &model[0])

	gl.BindVertexArray(r.quadVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)

	// Draw "RADAR" label
	r.DrawText(x+size/2-25, y+size+8, 1.2, "RADAR", [4]float32{0.4, 1.0, 0.6, 1.0})
}

// Cleanup releases resources
func (r *Renderer) Cleanup() {
	if r.quadVAO != 0 {
		gl.DeleteVertexArrays(1, &r.quadVAO)
	}
	if r.quadVBO != 0 {
		gl.DeleteBuffers(1, &r.quadVBO)
	}
	if r.cubeVAO != 0 {
		gl.DeleteVertexArrays(1, &r.cubeVAO)
	}
	if r.cubeVBO != 0 {
		gl.DeleteBuffers(1, &r.cubeVBO)
	}
	if r.shader != nil && r.shader.ID != 0 {
		gl.DeleteProgram(r.shader.ID)
	}
}

func (r *Renderer) createCubeMesh() {
	// Unit cube vertices with normals for 3D UI items
	vertices := []float32{
		// Positions        // Normals (approximate for simple shading)
		// Front face
		-0.5, -0.5, 0.5, 0, 0, 1,
		0.5, -0.5, 0.5, 0, 0, 1,
		0.5, 0.5, 0.5, 0, 0, 1,
		-0.5, -0.5, 0.5, 0, 0, 1,
		0.5, 0.5, 0.5, 0, 0, 1,
		-0.5, 0.5, 0.5, 0, 0, 1,

		// Back face (optional for UI?) - keep for completeness
		0.5, -0.5, -0.5, 0, 0, -1,
		-0.5, -0.5, -0.5, 0, 0, -1,
		-0.5, 0.5, -0.5, 0, 0, -1,
		0.5, -0.5, -0.5, 0, 0, -1,
		-0.5, 0.5, -0.5, 0, 0, -1,
		0.5, 0.5, -0.5, 0, 0, -1,

		// Top face
		-0.5, 0.5, 0.5, 0, 1, 0,
		0.5, 0.5, 0.5, 0, 1, 0,
		0.5, 0.5, -0.5, 0, 1, 0,
		-0.5, 0.5, 0.5, 0, 1, 0,
		0.5, 0.5, -0.5, 0, 1, 0,
		-0.5, 0.5, -0.5, 0, 1, 0,

		// Bottom face
		-0.5, -0.5, -0.5, 0, -1, 0,
		0.5, -0.5, -0.5, 0, -1, 0,
		0.5, -0.5, 0.5, 0, -1, 0,
		-0.5, -0.5, -0.5, 0, -1, 0,
		0.5, -0.5, 0.5, 0, -1, 0,
		-0.5, -0.5, 0.5, 0, -1, 0,

		// Right face
		0.5, -0.5, 0.5, 1, 0, 0,
		0.5, -0.5, -0.5, 1, 0, 0,
		0.5, 0.5, -0.5, 1, 0, 0,
		0.5, -0.5, 0.5, 1, 0, 0,
		0.5, 0.5, -0.5, 1, 0, 0,
		0.5, 0.5, 0.5, 1, 0, 0,

		// Left face
		-0.5, -0.5, -0.5, -1, 0, 0,
		-0.5, -0.5, 0.5, -1, 0, 0,
		-0.5, 0.5, 0.5, -1, 0, 0,
		-0.5, -0.5, -0.5, -1, 0, 0,
		-0.5, 0.5, 0.5, -1, 0, 0,
		-0.5, 0.5, -0.5, -1, 0, 0,
	}

	gl.GenVertexArrays(1, &r.cubeVAO)
	gl.GenBuffers(1, &r.cubeVBO)

	gl.BindVertexArray(r.cubeVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.cubeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Strides match renderer logic, but shader expects vec2 pos, vec2 UV?
	// Wait, the UI shader is 2D!
	// I CANNOT render 3D meshes with the current UI shader easily if it expects vec2 position.
	// Current UI Shader: layout(location = 0) in vec2 aPos; layout(location = 1) in vec2 aTexCoord;

	// I need to use a separate shader for 3D items OR modify the UI shader to support 3D attribute (ignoring z for 2D, or having a bool switch).
	// Easier: Just create a "UIShader3D" or similar, or just project 3D to 2D in code? No, let's update the shader.
	// Updating the shader... I can change attribute 0 to vec3.

	// For now, let's update the UI shader to take vec3 aPos to support proper 3D rotation.
	// We'll update createUIShader below.

	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 6*4, 0)
	gl.EnableVertexAttribArray(0)
	// Normal as generic attribute 1? UI shader uses aTexCoord at 1.
	// Let's rely on flat colors for now and fake lighting in shader or just use flat colors.
	// But I want good looking items.
	// Let's pass normals.
	gl.VertexAttribPointerWithOffset(2, 3, gl.FLOAT, false, 6*4, 3*4)
	gl.EnableVertexAttribArray(2) // Location 2 for normals

	gl.BindVertexArray(0)
}

// Render3DItemInBox renders a 3D item within a 2D box area
func (r *Renderer) Render3DItemInBox(x, y, size float32, item block.Type, color [3]float32) {
	if r.shader == nil {
		return
	}

	gl.UseProgram(r.shader.ID)

	// Enable depth test for 3D item
	gl.Enable(gl.DEPTH_TEST)
	gl.Clear(gl.DEPTH_BUFFER_BIT) // Clear depth for item overlay? No, might clear entire screen.
	// Just clearing the depth buffer area for the item is tricky.
	// Instead, just draw it on top with depth test enabled but cleared beforehand?
	// Actually, since UI is drawn last and on top (Ortho), 3D items inside it need to handle their own depth.
	// Let's just enable depth test and hope the z-buffer from world doesn't interfere (usually cleared before UI or UI uses different range).
	// UI typically uses no depth test.

	// We can clear just the depth buffer bit if we want these items to self-occlude perfectly.
	gl.Clear(gl.DEPTH_BUFFER_BIT)

	// Viewport logic? No, just transform.

	// Center of box
	cx := x + size/2
	cy := y + size/2

	// Matrix setup
	projection := mgl32.Ortho(0, float32(r.width), float32(r.height), 0, -1000, 1000) // Deep Z for 3D rotation
	projLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uProjection\x00"))
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

	// Model Matrix
	baseModel := mgl32.Translate3D(cx, cy, 0)
	// Scale to fit box
	scale := size * 0.4
	baseModel = baseModel.Mul4(mgl32.Scale3D(scale, scale, scale))

	// Rotate for isometric-ish view
	// Flip Y because Ortho Y is down/up?
	// Ortho is (0,0) top-left.
	baseModel = baseModel.Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(30)))
	baseModel = baseModel.Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(45)))

	// We need to implement the specific item logic again here.
	// Duplicating small logic from creature.go for separation.

	r.renderItemModel(item, baseModel, color)

	// Restore state
	gl.Disable(gl.DEPTH_TEST)
	gl.UseProgram(r.shader.ID) // Restore program if changed (it wasn't)

	// Re-set projection for 2D
	// (DrawRect handles its own uniforms so we are fine)
}

func (r *Renderer) renderItemModel(item block.Type, baseModel mgl32.Mat4, mainColor [3]float32) {
	handleColor := [3]float32{0.55, 0.35, 0.17}

	// Uniforms for light
	// Standard UI shader doesn't have lighting.
	// We need to use "uUseTexture" false + uColor.
	// To get 3D shading, we might need a dedicated shader or Hack it by altering color based on normal (if we had access to normal in frag shader).
	// Current fragment shader:
	// uniform vec4 uColor;
	// ...
	// fragColor = uColor;

	// To look good, we can just pass different uColor for different sides?
	// But we are drawing a cube VAO with many faces at once.
	// Without a lighting shader, it will look flat solid color.
	// We should update the UI shader to support basic lighting or normals!

	// Let's assume we update shader below.

	switch item {
	case block.Pickaxe:
		r.renderCubePart(baseModel, mgl32.Vec3{0.15, 1.4, 0.15}, mgl32.Vec3{0, 0, 0}, handleColor)   // Handle
		r.renderCubePart(baseModel, mgl32.Vec3{0.4, 0.25, 0.25}, mgl32.Vec3{0, 0.7, 0}, mainColor)   // Head center
		r.renderCubePart(baseModel, mgl32.Vec3{0.3, 0.2, 0.2}, mgl32.Vec3{-0.3, 0.65, 0}, mainColor) // L
		r.renderCubePart(baseModel, mgl32.Vec3{0.3, 0.2, 0.2}, mgl32.Vec3{0.3, 0.65, 0}, mainColor)  // R

	case block.Sword:
		r.renderCubePart(baseModel, mgl32.Vec3{0.15, 0.5, 0.15}, mgl32.Vec3{0, -0.6, 0}, handleColor)
		r.renderCubePart(baseModel, mgl32.Vec3{0.6, 0.1, 0.2}, mgl32.Vec3{0, -0.3, 0}, mainColor)
		r.renderCubePart(baseModel, mgl32.Vec3{0.2, 1.4, 0.1}, mgl32.Vec3{0, 0.5, 0}, mainColor)

	case block.Axe:
		r.renderCubePart(baseModel, mgl32.Vec3{0.15, 1.4, 0.15}, mgl32.Vec3{0, 0, 0}, handleColor)
		r.renderCubePart(baseModel, mgl32.Vec3{0.3, 0.3, 0.2}, mgl32.Vec3{0.1, 0.6, 0}, mainColor)
		r.renderCubePart(baseModel, mgl32.Vec3{0.2, 0.6, 0.2}, mgl32.Vec3{0.3, 0.6, 0}, mainColor)
		r.renderCubePart(baseModel, mgl32.Vec3{0.15, 0.2, 0.2}, mgl32.Vec3{-0.1, 0.6, 0}, mainColor)

	case block.Shovel:
		r.renderCubePart(baseModel, mgl32.Vec3{0.15, 1.4, 0.15}, mgl32.Vec3{0, 0.2, 0}, handleColor)
		r.renderCubePart(baseModel, mgl32.Vec3{0.5, 0.6, 0.1}, mgl32.Vec3{0, -0.7, 0}, mainColor)

	default:
		// Cube
		r.renderCubePart(baseModel, mgl32.Vec3{1, 1, 1}, mgl32.Vec3{0, 0, 0}, mainColor)
	}
}

func (r *Renderer) renderCubePart(baseModel mgl32.Mat4, size mgl32.Vec3, offset mgl32.Vec3, color [3]float32) {
	model := baseModel.Mul4(mgl32.Translate3D(offset.X(), offset.Y(), offset.Z()))
	model = model.Mul4(mgl32.Scale3D(size.X(), size.Y(), size.Z()))

	// Set uniforms
	// Projection is set in Render3DItemInBox, so we skip it here for efficiency
	// (Assuming we only call this from there)

	modLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uModel\x00"))
	gl.UniformMatrix4fv(modLoc, 1, false, &model[0])

	colLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uColor\x00"))
	gl.Uniform4f(colLoc, color[0], color[1], color[2], 1.0)

	useTexLoc := gl.GetUniformLocation(r.shader.ID, gl.Str("uUseTexture\x00"))
	gl.Uniform1i(useTexLoc, 0)

	// Draw
	gl.BindVertexArray(r.cubeVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 36)
	gl.BindVertexArray(0)
}

func (r *Renderer) createQuadMesh() {
	// Unit quad vertices (2D positions + UVs)
	vertices := []float32{
		// Pos      // UV
		0, 0, 0, 0,
		1, 0, 1, 0,
		1, 1, 1, 1,
		0, 0, 0, 0,
		1, 1, 1, 1,
		0, 1, 0, 1,
	}

	gl.GenVertexArrays(1, &r.quadVAO)
	gl.GenBuffers(1, &r.quadVBO)

	gl.BindVertexArray(r.quadVAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, r.quadVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Position
	gl.VertexAttribPointerWithOffset(0, 2, gl.FLOAT, false, 4*4, 0)
	gl.EnableVertexAttribArray(0)
	// UV
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, 4*4, 2*4)
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

func createUIShader() (*UIShader, error) {
	vertexSource := `
#version 410 core
layout(location = 0) in vec3 aPos; // Changed to vec3 to support 3D
layout(location = 1) in vec2 aTexCoord;
layout(location = 2) in vec3 aNormal; // Added normal

uniform mat4 uProjection;
uniform mat4 uModel;

out vec2 vTexCoord;
out vec3 vNormal; 

void main() {
    // Standard UI transform
    gl_Position = uProjection * uModel * vec4(aPos, 1.0);
    vTexCoord = aTexCoord;
    
    // Transform normal
    vNormal = mat3(transpose(inverse(uModel))) * aNormal;
}
` + "\x00"

	fragmentSource := `
#version 410 core
in vec2 vTexCoord;
in vec3 vNormal;

uniform vec4 uColor;
uniform sampler2D uTexture;
uniform bool uUseTexture;
uniform vec4 uUVBounds; 

out vec4 fragColor;

void main() {
    if (uUseTexture) {
        float u = uUVBounds.x + vTexCoord.x * (uUVBounds.z - uUVBounds.x);
        float v = uUVBounds.y + vTexCoord.y * (uUVBounds.w - uUVBounds.y);
        vec4 texColor = texture(uTexture, vec2(u, v));
        if (texColor.a < 0.1) discard;
        fragColor = vec4(1.0, 1.0, 1.0, texColor.r) * uColor;
    } else {
        // Simple lighting for 3D items
        if (length(vNormal) > 0.1) {
            vec3 lightDir = normalize(vec3(0.5, 0.8, 1.0));
            float diff = max(dot(normalize(vNormal), lightDir), 0.3); // Ambient 0.3
            fragColor = vec4(uColor.rgb * diff, uColor.a);
        } else {
            fragColor = uColor;
        }
    }
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
