package render

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// TextureManager manages OpenGL textures
type TextureManager struct {
	BlockTextureArray uint32
	TextureSize       int32
}

// NewTextureManager creates a new texture manager
func NewTextureManager() *TextureManager {
	return &TextureManager{
		TextureSize: 16, // Default to 16x16 pixels per block
	}
}

// LoadBlockTextures loads a list of image files from disk into a 2D Texture Array
func (tm *TextureManager) LoadBlockTextures(files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files to load")
	}

	layerCount := int32(len(files))
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D_ARRAY, textureID)

	// Allocate storage for the texture array
	// Mipmap levels: log2(16) = 4
	mipLevels := int32(1)
	size := tm.TextureSize
	for size > 1 {
		size /= 2
		mipLevels++
	}

	gl.TexStorage3D(gl.TEXTURE_2D_ARRAY, mipLevels, gl.RGBA8, tm.TextureSize, tm.TextureSize, layerCount)

	// Upload images
	for i, file := range files {
		img, err := loadExactImage(file, int(tm.TextureSize))
		if err != nil {
			fmt.Printf("Warning: Failed to load texture %s: %v. Using magenta placeholder.\n", file, err)
			img = createPlaceholderImage(int(tm.TextureSize))
		}

		rgba := imageToRGBA(img)
		gl.TexSubImage3D(gl.TEXTURE_2D_ARRAY, 0, 0, 0, int32(i), tm.TextureSize, tm.TextureSize, 1, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
	}

	// Generate mipmaps
	gl.GenerateMipmap(gl.TEXTURE_2D_ARRAY)

	// Set parameters
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MIN_FILTER, gl.NEAREST_MIPMAP_LINEAR) // Pixelated look
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_T, gl.REPEAT)

	tm.BlockTextureArray = textureID
	return nil
}

// LoadBlockTexturesFromEmbed loads textures from an embedded filesystem (go:embed)
// This allows the game to be distributed as a single executable
func (tm *TextureManager) LoadBlockTexturesFromEmbed(files []string, fs embed.FS) error {
	if len(files) == 0 {
		return fmt.Errorf("no files to load")
	}

	layerCount := int32(len(files))
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D_ARRAY, textureID)

	// Allocate storage for the texture array
	mipLevels := int32(1)
	size := tm.TextureSize
	for size > 1 {
		size /= 2
		mipLevels++
	}

	gl.TexStorage3D(gl.TEXTURE_2D_ARRAY, mipLevels, gl.RGBA8, tm.TextureSize, tm.TextureSize, layerCount)

	// Upload images from embedded filesystem
	for i, file := range files {
		img, err := loadImageFromEmbed(fs, file, int(tm.TextureSize))
		if err != nil {
			fmt.Printf("Warning: Failed to load embedded texture %s: %v. Using magenta placeholder.\n", file, err)
			img = createPlaceholderImage(int(tm.TextureSize))
		}

		rgba := imageToRGBA(img)
		gl.TexSubImage3D(gl.TEXTURE_2D_ARRAY, 0, 0, 0, int32(i), tm.TextureSize, tm.TextureSize, 1, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
	}

	// Generate mipmaps
	gl.GenerateMipmap(gl.TEXTURE_2D_ARRAY)

	// Set parameters
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MIN_FILTER, gl.NEAREST_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_T, gl.REPEAT)

	tm.BlockTextureArray = textureID
	return nil
}

// BindBlockTextures binds the texture array to a texture unit
func (tm *TextureManager) BindBlockTextures(unit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + unit)
	gl.BindTexture(gl.TEXTURE_2D_ARRAY, tm.BlockTextureArray)
}

// Cleanup releases resources
func (tm *TextureManager) Cleanup() {
	if tm.BlockTextureArray != 0 {
		gl.DeleteTextures(1, &tm.BlockTextureArray)
		tm.BlockTextureArray = 0
	}
}

// Helper functions

func loadExactImage(path string, size int) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	// Create a new RGBA image of the desired size
	dst := image.NewRGBA(image.Rect(0, 0, size, size))

	// Draw (resize/crop if needed - for now simple draw)
	// In a real engine we might want high-quality scaling, but for voxels exact pixel art is usually preferred.
	// If source is different size, this might look bad without scaling.
	// Assuming textures are correct size for now.
	draw.Draw(dst, dst.Bounds(), img, image.Point{}, draw.Src)

	return dst, nil
}

// loadImageFromEmbed loads an image from an embedded filesystem
func loadImageFromEmbed(fs embed.FS, path string, size int) (image.Image, error) {
	data, err := fs.ReadFile(path)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// Create a new RGBA image of the desired size
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(dst, dst.Bounds(), img, image.Point{}, draw.Src)

	return dst, nil
}

func imageToRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	return rgba
}

func createPlaceholderImage(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	// Magenta
	col := image.NewUniform(color.RGBA{255, 0, 255, 255})
	draw.Draw(img, img.Bounds(), col, image.Point{}, draw.Src)
	// Set color to Magenta (255, 0, 255)
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = 255   // R
		img.Pix[i+1] = 0   // G
		img.Pix[i+2] = 255 // B
		img.Pix[i+3] = 255 // A
	}
	return img
}
