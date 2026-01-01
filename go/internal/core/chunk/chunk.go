// Package chunk manages world chunks for the voxel engine
package chunk

import (
	"voxelgame/internal/core/block"
)

// Size constants for chunks
const (
	Size   = 16 // Width and depth of a chunk
	Height = 64 // Height of a chunk
)

// Chunk represents a 16x64x16 section of the world
type Chunk struct {
	// Position in chunk coordinates
	CX, CZ int32

	// Block data stored in a flat array for cache efficiency
	// Index = x + y*Size + z*Size*Height
	Data []block.Type

	// Height map for quick surface lookups
	HeightMap []uint8

	// Flags
	IsGenerated bool
	IsDirty     bool

	// Statistics
	SolidBlockCount int

	// OpenGL handles (set by renderer)
	VAO         uint32
	VBO         uint32
	EBO         uint32
	VertexCount int32
}

// New creates a new empty chunk at the given chunk coordinates
func New(cx, cz int32) *Chunk {
	return &Chunk{
		CX:        cx,
		CZ:        cz,
		Data:      make([]block.Type, Size*Height*Size),
		HeightMap: make([]uint8, Size*Size),
		IsDirty:   true,
	}
}

// ID returns a unique string identifier for this chunk
func (c *Chunk) ID() string {
	return ChunkID(int(c.CX), int(c.CZ))
}

// ChunkID creates a chunk ID string from coordinates
func ChunkID(cx, cz int) string {
	// Simple concatenation for map keys
	return string(rune(cx)) + "," + string(rune(cz))
}

// getIndex converts local coordinates to array index
func (c *Chunk) getIndex(lx, ly, lz int) int {
	return lx + ly*Size + lz*Size*Height
}

// GetBlock returns the block type at local coordinates
func (c *Chunk) GetBlock(lx, ly, lz int) block.Type {
	if lx < 0 || lx >= Size || lz < 0 || lz >= Size || ly < 0 || ly >= Height {
		return block.Air
	}
	return c.Data[c.getIndex(lx, ly, lz)]
}

// SetBlock sets the block type at local coordinates
// Returns true if the block was changed
func (c *Chunk) SetBlock(lx, ly, lz int, t block.Type) bool {
	if lx < 0 || lx >= Size || lz < 0 || lz >= Size || ly < 0 || ly >= Height {
		return false
	}

	idx := c.getIndex(lx, ly, lz)
	oldType := c.Data[idx]

	if oldType == t {
		return false
	}

	c.Data[idx] = t
	c.IsDirty = true

	// Update solid block count
	if oldType == block.Air && t != block.Air {
		c.SolidBlockCount++
	} else if oldType != block.Air && t == block.Air {
		c.SolidBlockCount--
	}

	// Update height map
	if t != block.Air {
		hmIdx := lx + lz*Size
		if ly > int(c.HeightMap[hmIdx]) {
			c.HeightMap[hmIdx] = uint8(ly)
		}
	}

	return true
}

// GetHeight returns the highest block at local x,z
func (c *Chunk) GetHeight(lx, lz int) int {
	if lx < 0 || lx >= Size || lz < 0 || lz >= Size {
		return 0
	}
	return int(c.HeightMap[lx+lz*Size])
}

// ForEachSolidBlock iterates over all non-air blocks
func (c *Chunk) ForEachSolidBlock(fn func(lx, ly, lz int, t block.Type)) {
	for z := 0; z < Size; z++ {
		for y := 0; y < Height; y++ {
			for x := 0; x < Size; x++ {
				t := c.GetBlock(x, y, z)
				if t != block.Air {
					fn(x, y, z, t)
				}
			}
		}
	}
}

// IsExposed returns true if a block has any exposed faces
func (c *Chunk) IsExposed(lx, ly, lz int) bool {
	return c.GetBlock(lx-1, ly, lz) == block.Air ||
		c.GetBlock(lx+1, ly, lz) == block.Air ||
		c.GetBlock(lx, ly-1, lz) == block.Air ||
		c.GetBlock(lx, ly+1, lz) == block.Air ||
		c.GetBlock(lx, ly, lz-1) == block.Air ||
		c.GetBlock(lx, ly, lz+1) == block.Air
}

// VisibleFaces returns which faces of a block are visible
type VisibleFaces struct {
	Top, Bottom, Left, Right, Front, Back bool
}

// GetVisibleFaces returns which faces of a block are visible
func (c *Chunk) GetVisibleFaces(lx, ly, lz int) VisibleFaces {
	return VisibleFaces{
		Top:    ly == Height-1 || c.GetBlock(lx, ly+1, lz) == block.Air,
		Bottom: ly == 0 || c.GetBlock(lx, ly-1, lz) == block.Air,
		Left:   lx == 0 || c.GetBlock(lx-1, ly, lz) == block.Air,
		Right:  lx == Size-1 || c.GetBlock(lx+1, ly, lz) == block.Air,
		Front:  lz == Size-1 || c.GetBlock(lx, ly, lz+1) == block.Air,
		Back:   lz == 0 || c.GetBlock(lx, ly, lz-1) == block.Air,
	}
}

// Dispose cleans up OpenGL resources
func (c *Chunk) Dispose() {
	// OpenGL cleanup will be handled by the renderer
	c.VAO = 0
	c.VBO = 0
	c.EBO = 0
	c.VertexCount = 0
}

// Serialize converts the chunk to a serializable format
type SerializedChunk struct {
	CX        int32   `json:"cx"`
	CZ        int32   `json:"cz"`
	Data      []uint8 `json:"data"`
	HeightMap []uint8 `json:"heightMap"`
}

// Serialize returns the chunk data for saving
func (c *Chunk) Serialize() SerializedChunk {
	data := make([]uint8, len(c.Data))
	for i, b := range c.Data {
		data[i] = uint8(b)
	}

	return SerializedChunk{
		CX:        c.CX,
		CZ:        c.CZ,
		Data:      data,
		HeightMap: c.HeightMap,
	}
}

// Deserialize creates a chunk from serialized data
func Deserialize(s SerializedChunk) *Chunk {
	c := New(s.CX, s.CZ)

	for i, b := range s.Data {
		c.Data[i] = block.Type(b)
	}
	copy(c.HeightMap, s.HeightMap)

	c.IsGenerated = true
	c.IsDirty = true

	return c
}
