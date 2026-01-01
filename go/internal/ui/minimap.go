package ui

import (
	"image"
	"image/color"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Minimap displays a top-down view of the world
type Minimap struct {
	textureID uint32
	size      int // Texture size (e.g., 128)
	pixels    *image.RGBA
	scale     float64 // World blocks per pixel
}

// NewMinimap creates a new minimap
func NewMinimap(size int) *Minimap {
	m := &Minimap{
		size:   size,
		pixels: image.NewRGBA(image.Rect(0, 0, size, size)),
		scale:  1.0, // 1 block per pixel
	}
	m.initTexture()
	return m
}

func (m *Minimap) initTexture() {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	m.textureID = texture
}

// Update updates the minimap texture based on world data
func (m *Minimap) Update(playerPos mgl32.Vec3, getBiome func(x, z int) string, getHeight func(x, z int) int, creatures []mgl32.Vec3) {
	px := int(playerPos.X())
	pz := int(playerPos.Z())
	offset := m.size / 2

	// Render terrain
	for z := 0; z < m.size; z++ {
		for x := 0; x < m.size; x++ {
			wx := px + (x - offset)
			wz := pz + (z - offset)

			biome := getBiome(wx, wz)
			height := getHeight(wx, wz)

			var c color.RGBA

			// Simple biome colors
			switch biome {
			case "plains":
				c = color.RGBA{100, 200, 100, 255} // Green
			case "desert":
				c = color.RGBA{220, 200, 150, 255} // Sand
			case "snow":
				c = color.RGBA{240, 240, 255, 255} // White
			case "forest":
				c = color.RGBA{34, 139, 34, 255} // Forest Green
			case "mountains":
				c = color.RGBA{128, 128, 128, 255} // Gray
			default:
				c = color.RGBA{100, 100, 100, 255}
			}

			// Height shading (darker if lower)
			diff := height - int(playerPos.Y())
			shade := 1.0
			if diff < -10 {
				shade = 0.7
			} else if diff > 10 {
				shade = 1.2
			}

			// Water check (simplified, assume base logic)
			if height < 12 && biome != "desert" { // SeaLevel 12
				c = color.RGBA{50, 50, 200, 255} // Blue
			} else {
				c.R = uint8(float64(c.R) * shade)
				c.G = uint8(float64(c.G) * shade)
				c.B = uint8(float64(c.B) * shade)
			}

			m.pixels.Set(x, z, c)
		}
	}

	// Render creatures
	for _, pos := range creatures {
		cx := int(pos.X()) - px + offset
		cz := int(pos.Z()) - pz + offset

		if cx >= 0 && cx < m.size && cz >= 0 && cz < m.size {
			// Red dot 3x3
			col := color.RGBA{255, 0, 0, 255}
			m.pixels.Set(cx, cz, col)
			m.pixels.Set(cx+1, cz, col)
			m.pixels.Set(cx-1, cz, col)
			m.pixels.Set(cx, cz+1, col)
			m.pixels.Set(cx, cz-1, col)
		}
	}

	// Player center (White)
	center := color.RGBA{255, 255, 255, 255}
	m.pixels.Set(offset, offset, center)
	m.pixels.Set(offset+1, offset, center)
	m.pixels.Set(offset-1, offset, center)
	m.pixels.Set(offset, offset+1, center)
	m.pixels.Set(offset, offset-1, center)

	// Upload
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

func (m *Minimap) Bind() {
	gl.BindTexture(gl.TEXTURE_2D, m.textureID)
}

func (m *Minimap) GetTextureID() uint32 {
	return m.textureID
}
