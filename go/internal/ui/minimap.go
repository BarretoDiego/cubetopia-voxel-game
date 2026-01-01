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
	pz := int(playerPos.Z())
	center := m.size / 2
	radius := m.size / 2

	// Update scan angle
	m.updateCount++
	m.scanAngle = float64(m.updateCount) * 0.05 // Sweep speed
	if m.scanAngle > 2*math.Pi {
		m.scanAngle -= 2 * math.Pi
		m.updateCount = 0
	}

	// Clear to dark radar green/black
	for y := 0; y < m.size; y++ {
		for x := 0; x < m.size; x++ {
			dx := x - center
			dy := y - center
			dist := math.Sqrt(float64(dx*dx + dy*dy))

			if dist > float64(radius) {
				// Outside circle - transparent
				m.pixels.Set(x, y, color.RGBA{0, 0, 0, 0})
			} else {
				// Dark radar background
				m.pixels.Set(x, y, color.RGBA{0, 15, 8, 255})
			}
		}
	}

	// Render terrain with radar color scheme
	for y := 0; y < m.size; y++ {
		for x := 0; x < m.size; x++ {
			dx := x - center
			dy := y - center
			dist := math.Sqrt(float64(dx*dx + dy*dy))

			if dist > float64(radius)-2 {
				continue // Skip outside circle
			}

			// World coordinates
			wx := px + int(float64(dx)*m.scale)
			wz := pz + int(float64(dy)*m.scale)

			biome := getBiome(wx, wz)
			height := getHeight(wx, wz)

			var c color.RGBA

			// Radar-style monochrome green colors
			switch biome {
			case "plains":
				c = color.RGBA{30, 80, 50, 255}
			case "desert":
				c = color.RGBA{60, 70, 40, 255}
			case "snow":
				c = color.RGBA{80, 100, 90, 255}
			case "forest":
				c = color.RGBA{20, 60, 35, 255}
			case "mountains":
				c = color.RGBA{50, 60, 55, 255}
			default:
				c = color.RGBA{25, 50, 35, 255}
			}

			// Height shading
			diff := height - int(playerPos.Y())
			shade := 1.0
			if diff < -10 {
				shade = 0.6
			} else if diff > 10 {
				shade = 1.3
			}

			// Water
			if height < 12 && biome != "desert" {
				c = color.RGBA{20, 40, 80, 255}
			} else {
				c.R = uint8(clamp(float64(c.R)*shade, 0, 255))
				c.G = uint8(clamp(float64(c.G)*shade, 0, 255))
				c.B = uint8(clamp(float64(c.B)*shade, 0, 255))
			}

			m.pixels.Set(x, y, c)
		}
	}

	// Draw radar grid rings
	gridColor := color.RGBA{0, 100, 60, 100}
	for _, ratio := range []float64{0.25, 0.5, 0.75, 1.0} {
		r := float64(radius) * ratio
		for angle := 0.0; angle < 2*math.Pi; angle += 0.02 {
			rx := center + int(math.Cos(angle)*r)
			ry := center + int(math.Sin(angle)*r)
			if rx >= 0 && rx < m.size && ry >= 0 && ry < m.size {
				m.pixels.Set(rx, ry, gridColor)
			}
		}
	}

	// Draw crosshair
	for i := 0; i < m.size; i++ {
		if i >= 0 && i < m.size {
			// Vertical
			if math.Abs(float64(i-center)) < float64(radius) {
				m.pixels.Set(center, i, gridColor)
			}
			// Horizontal
			if math.Abs(float64(i-center)) < float64(radius) {
				m.pixels.Set(i, center, gridColor)
			}
		}
	}

	// Draw scan sweep line
	sweepColor := color.RGBA{100, 255, 150, 200}
	for r := 0.0; r < float64(radius); r += 0.5 {
		sx := center + int(math.Cos(m.scanAngle)*r)
		sy := center + int(math.Sin(m.scanAngle)*r)
		if sx >= 0 && sx < m.size && sy >= 0 && sy < m.size {
			m.pixels.Set(sx, sy, sweepColor)
		}
	}

	// Draw creatures as bright blips
	for _, pos := range creatures {
		cx := int((pos.X()-float32(px))/float32(m.scale)) + center
		cz := int((pos.Z()-float32(pz))/float32(m.scale)) + center

		// Check if within radar circle
		dx := cx - center
		dz := cz - center
		dist := math.Sqrt(float64(dx*dx + dz*dz))

		if dist < float64(radius)-4 && cx >= 2 && cx < m.size-2 && cz >= 2 && cz < m.size-2 {
			// Calculate angle to creature for fade effect
			creatureAngle := math.Atan2(float64(dz), float64(dx))
			angleDiff := m.scanAngle - creatureAngle
			if angleDiff < 0 {
				angleDiff += 2 * math.Pi
			}

			// Brightness based on scan recency
			brightness := uint8(255)
			if angleDiff > 0.5 {
				brightness = uint8(clamp(255*(1-angleDiff/(2*math.Pi)), 80, 255))
			}

			col := color.RGBA{brightness, 50, 50, 255} // Red blip
			m.pixels.Set(cx, cz, col)
			m.pixels.Set(cx+1, cz, col)
			m.pixels.Set(cx-1, cz, col)
			m.pixels.Set(cx, cz+1, col)
			m.pixels.Set(cx, cz-1, col)
		}
	}

	// Player center (bright white/cyan)
	playerColor := color.RGBA{200, 255, 220, 255}
	m.pixels.Set(center, center, playerColor)
	m.pixels.Set(center+1, center, playerColor)
	m.pixels.Set(center-1, center, playerColor)
	m.pixels.Set(center, center+1, playerColor)
	m.pixels.Set(center, center-1, playerColor)

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
