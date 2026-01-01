// Package assets provides embedded game assets (shaders, textures)
// This allows the game to be distributed as a single executable
package assets

import (
	"embed"
	"io/fs"
)

//go:embed shaders/*.vert shaders/*.frag textures/*.png
var embeddedFS embed.FS

// FS returns the embedded filesystem containing all assets
func FS() embed.FS {
	return embeddedFS
}

// ReadFile reads a file from the embedded filesystem
// Path should be relative to the assets directory (e.g., "shaders/voxel.vert")
func ReadFile(path string) ([]byte, error) {
	return embeddedFS.ReadFile(path)
}

// ReadTexture reads a texture file from embedded assets
func ReadTexture(name string) ([]byte, error) {
	return embeddedFS.ReadFile("textures/" + name)
}

// ReadShader reads a shader file from embedded assets
func ReadShader(name string) ([]byte, error) {
	return embeddedFS.ReadFile("shaders/" + name)
}

// ListShaders returns all shader files
func ListShaders() ([]string, error) {
	var files []string
	err := fs.WalkDir(embeddedFS, "shaders", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// ListTextures returns all texture files
func ListTextures() ([]string, error) {
	var files []string
	err := fs.WalkDir(embeddedFS, "textures", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
