package ui

import (
	"image"
	"image/color"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Minimap displays a top-down view of the world as a radar
type Minimap struct {
	textureID   uint32
	size        int // Texture size (e.g., 256)
	pixels      *image.RGBA
	scale       float64 // World blocks per pixel
	scanAngle   float64 // Current sweep angle in radians
	updateCount int     // Frame counter for animation
}

// NewMinimap creates a new minimap
func NewMinimap(size int) *Minimap {
	m := &Minimap{
		size:      size,
		pixels:    image.NewRGBA(image.Rect(0, 0, size, size)),
		scale:     0.5, // 0.5 block per pixel (larger area)
		scanAngle: 0,
	}
	m.initTexture()
	return m
}

func (m *Minimap) initTexture() {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	m.textureID = texture
}

// Update updates the minimap texture based on world data
func (m *Minimap) Update(playerPos mgl32.Vec3, getBiome func(x, z int) string, getHeight func(x, z int) int, creatures []mgl32.Vec3) {
	px := int(playerPos.X())
	py := int(playerPos.Y())
	pz := int(playerPos.Z())
	center := m.size / 2
	radius := m.size / 2

	// Update scan angle (sweep line still looks cool as an overlay)
	m.updateCount++
	m.scanAngle = float64(m.updateCount) * 0.05
	if m.scanAngle > 2*math.Pi {
		m.scanAngle -= 2 * math.Pi
		m.updateCount = 0
	}

	// 1. Clear background
	for y := 0; y < m.size; y++ {
		for x := 0; x < m.size; x++ {
			dx := x - center
			dy := y - center
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > float64(radius) {
				m.pixels.Set(x, y, color.RGBA{0, 0, 0, 0})
			} else {
				m.pixels.Set(x, y, color.RGBA{0, 10, 5, 255})
			}
		}
	}

	// 2. Render Isometric Terrain
	// We'll scan a square area around the player and project it
	viewRange := 24 // Number of blocks to show
	isoScaleX := 4.0
	isoScaleY := 2.0
	heightScale := 3.0

	// Helper to set pixel if within radar circle
	setPixel := func(x, y int, c color.RGBA) {
		if x < 0 || x >= m.size || y < 0 || y >= m.size {
			return
		}
		dx := x - center
		dy := y - center
		if dx*dx+dy*dy < radius*radius {
			m.pixels.Set(x, y, c)
		}
	}

	// Draw blocks back-to-front
	// Z goes from min to max, X goes from min to max
	for dz := -viewRange; dz <= viewRange; dz++ {
		for dx := -viewRange; dx <= viewRange; dx++ {
			wx := px + dx
			wz := pz + dz

			biome := getBiome(wx, wz)
			h := getHeight(wx, wz)

			// Isometric projection
			// ScreenX = (dx - dz) * (isoScaleX / 2) + center
			// ScreenY = (dx + dz) * (isoScaleY / 2) - (h - py) * heightScale + center
			sx := float64(dx-dz)*(isoScaleX/2) + float64(center)
			sy := float64(dx+dz)*(isoScaleY/2) - float64(h-py)*heightScale + float64(center)

			// Get base color
			var c color.RGBA
			switch biome {
			case "plains":
				c = color.RGBA{45, 120, 60, 255}
			case "desert":
				c = color.RGBA{180, 160, 80, 255}
			case "snow":
				c = color.RGBA{200, 220, 255, 255}
			case "forest":
				c = color.RGBA{30, 90, 40, 255}
			case "mountains":
				c = color.RGBA{100, 110, 115, 255}
			default:
				c = color.RGBA{50, 100, 60, 255}
			}

			// Water
			if h < 12 && biome != "desert" {
				c = color.RGBA{40, 80, 200, 200}
				h = 12 // Level water surface
			}

			// Height shading
			shade := 1.0 + float64(h-py)*0.05
			c.R = uint8(clamp(float64(c.R)*shade, 0, 255))
			c.G = uint8(clamp(float64(c.G)*shade, 0, 255))
			c.B = uint8(clamp(float64(c.B)*shade, 0, 255))

			// Draw a small block shape (diamond)
			ix, iy := int(sx), int(sy)
			setPixel(ix, iy, c)   // Center
			setPixel(ix-1, iy, c) // Left
			setPixel(ix+1, iy, c) // Right
			setPixel(ix, iy-1, c) // Top
			setPixel(ix, iy+1, c) // Bottom
			setPixel(ix-2, iy, c) // Far Left
			setPixel(ix+2, iy, c) // Far Right

			// Optional: draw "walls" for elevation depth
			wallColor := color.RGBA{uint8(float64(c.R) * 0.7), uint8(float64(c.G) * 0.7), uint8(float64(c.B) * 0.7), 255}
			for wh := 1; wh < 3; wh++ {
				setPixel(ix, iy+wh+1, wallColor)
				setPixel(ix-1, iy+wh, wallColor)
				setPixel(ix+1, iy+wh, wallColor)
			}
		}
	}

	// 3. Draw grid and trim
	// Rings
	for _, ratio := range []float64{0.5, 1.0} {
		r := float64(radius) * ratio
		for angle := 0.0; angle < 2*math.Pi; angle += 0.01 {
			rx := center + int(math.Cos(angle)*r)
			ry := center + int(math.Sin(angle)*r)
			if rx >= 0 && rx < m.size && ry >= 0 && ry < m.size {
				m.pixels.Set(rx, ry, color.RGBA{0, 255, 100, 80})
			}
		}
	}

	// Crosshair
	for i := center - 5; i <= center+5; i++ {
		m.pixels.Set(center, i, color.RGBA{255, 255, 255, 150})
		m.pixels.Set(i, center, color.RGBA{255, 255, 255, 150})
	}

	// 4. Draw scan sweep
	sweepColor := color.RGBA{100, 255, 150, 60}
	for r := 0.0; r < float64(radius); r += 1.0 {
		sx := center + int(math.Cos(m.scanAngle)*r)
		sy := center + int(math.Sin(m.scanAngle)*r)
		if sx >= 0 && sx < m.size && sy >= 0 && sy < m.size {
			m.pixels.Set(sx, sy, sweepColor)
		}
	}

	// 5. Draw creatures
	for _, pos := range creatures {
		dx := int(pos.X()) - px
		dz := int(pos.Z()) - pz
		h := getHeight(int(pos.X()), int(pos.Z()))

		sx := float64(dx-dz)*(isoScaleX/2) + float64(center)
		sy := float64(dx+dz)*(isoScaleY/2) - float64(h-py)*heightScale + float64(center)

		creatureAngle := math.Atan2(float64(dz), float64(dx))
		angleDiff := m.scanAngle - creatureAngle
		if angleDiff < 0 {
			angleDiff += 2 * math.Pi
		}

		brightness := uint8(255)
		if angleDiff > 0.5 {
			brightness = uint8(clamp(255*(1-angleDiff/(2*math.Pi)), 100, 255))
		}

		col := color.RGBA{brightness, 20, 20, 255}
		ix, iy := int(sx), int(sy)
		setPixel(ix, iy, col)
		setPixel(ix+1, iy, col)
		setPixel(ix-1, iy, col)
		setPixel(ix, iy+1, col)
		setPixel(ix, iy-1, col)
	}

	// Player center handled by crosshair mostly, but add a dot
	playerColor := color.RGBA{255, 255, 255, 255}
	m.pixels.Set(center, center, playerColor)

	// Upload texture
	gl.BindTexture(gl.TEXTURE_2D, m.textureID)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(m.size),
		int32(m.size),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(m.pixels.Pix),
	)
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func (m *Minimap) Bind() {
	gl.BindTexture(gl.TEXTURE_2D, m.textureID)
}

func (m *Minimap) GetTextureID() uint32 {
	return m.textureID
}
