// Package chunk provides mesh generation for voxel chunks
package chunk

import (
	"voxelgame/internal/core/block"
)

// Vertex data layout for OpenGL
// Position (3) + Normal (3) + Color (3) + AO (1) + TexCoord (2) + MaterialID (1) = 13 floats
const VertexSize = 13

// Standard UV coordinates for a quad
var faceUVs = [4][2]float32{
	{0, 0},
	{1, 0},
	{1, 1},
	{0, 1},
}

// Face vertices for a cube
var faceVertices = map[string][4][3]float32{
	"top":    {{0, 1, 0}, {1, 1, 0}, {1, 1, 1}, {0, 1, 1}},
	"bottom": {{0, 0, 1}, {1, 0, 1}, {1, 0, 0}, {0, 0, 0}},
	"front":  {{0, 0, 1}, {0, 1, 1}, {1, 1, 1}, {1, 0, 1}},
	"back":   {{1, 0, 0}, {1, 1, 0}, {0, 1, 0}, {0, 0, 0}},
	"left":   {{0, 0, 0}, {0, 1, 0}, {0, 1, 1}, {0, 0, 1}},
	"right":  {{1, 0, 1}, {1, 1, 1}, {1, 1, 0}, {1, 0, 0}},
}

// Face normals
var faceNormals = map[string][3]float32{
	"top":    {0, 1, 0},
	"bottom": {0, -1, 0},
	"front":  {0, 0, 1},
	"back":   {0, 0, -1},
	"left":   {-1, 0, 0},
	"right":  {1, 0, 0},
}

// Neighbor offsets for each face
var neighborOffsets = [][3]int{
	{0, 1, 0},  // top
	{0, -1, 0}, // bottom
	{0, 0, 1},  // front
	{0, 0, -1}, // back
	{-1, 0, 0}, // left
	{1, 0, 0},  // right
}

// Face names in order
var faceNames = []string{"top", "bottom", "front", "back", "left", "right"}

// MeshData contains the generated mesh data for a chunk
type MeshData struct {
	Vertices    []float32
	Indices     []uint32
	VertexCount int
	IndexCount  int
}

// Mesher generates optimized meshes for chunks
type Mesher struct {
	// Buffers for building mesh
	vertices []float32
	indices  []uint32
}

// NewMesher creates a new chunk mesher
func NewMesher() *Mesher {
	return &Mesher{
		vertices: make([]float32, 0, 65536),
		indices:  make([]uint32, 0, 65536),
	}
}

// BlockGetter is a function that returns a block at world coordinates
type BlockGetter func(wx, wy, wz int) block.Type

// GenerateMesh generates mesh data for a chunk
func (m *Mesher) GenerateMesh(c *Chunk, getBlock BlockGetter) *MeshData {
	m.resetBuffers()

	worldOffsetX := int(c.CX) * Size
	worldOffsetZ := int(c.CZ) * Size

	// Iterate over all blocks
	for z := 0; z < Size; z++ {
		for y := 0; y < Height; y++ {
			for x := 0; x < Size; x++ {
				blockType := c.GetBlock(x, y, z)

				if blockType == block.Air {
					continue
				}

				blockDef := block.GetDefinition(blockType)
				worldX := worldOffsetX + x
				worldZ := worldOffsetZ + z

				// Check each face
				m.addVisibleFaces(
					x, y, z,
					worldX, y, worldZ,
					blockType, blockDef,
					c, getBlock,
				)
			}
		}
	}

	if len(m.vertices) == 0 {
		return nil
	}

	return &MeshData{
		Vertices:    append([]float32{}, m.vertices...),
		Indices:     append([]uint32{}, m.indices...),
		VertexCount: len(m.vertices) / VertexSize,
		IndexCount:  len(m.indices),
	}
}

// addVisibleFaces adds visible faces of a block to the mesh
func (m *Mesher) addVisibleFaces(
	lx, ly, lz int,
	wx, wy, wz int,
	blockType block.Type,
	blockDef block.Definition,
	c *Chunk,
	getBlock BlockGetter,
) {
	for i, face := range faceNames {
		offset := neighborOffsets[i]
		neighborType := getBlock(wx+offset[0], wy+offset[1], wz+offset[2])

		// Only add face if neighbor is air or transparent (and different type)
		neighborDef := block.GetDefinition(neighborType)

		if neighborType == block.Air ||
			(neighborDef.Transparent && neighborType != blockType) {
			m.addFace(face, float32(wx), float32(wy), float32(wz), blockDef, c, lx, ly, lz, getBlock)
		}
	}
}

// addFace adds a single face to the mesh buffers
func (m *Mesher) addFace(
	face string,
	x, y, z float32,
	blockDef block.Definition,
	c *Chunk,
	lx, ly, lz int,
	getBlock BlockGetter,
) {
	vertices := faceVertices[face]
	normal := faceNormals[face]
	baseIndex := uint32(len(m.vertices) / VertexSize)

	color := blockDef.Color
	materialID := float32(blockDef.Material)

	// Add 4 vertices for the face
	for i := 0; i < 4; i++ {
		vx := vertices[i][0]
		vy := vertices[i][1]
		vz := vertices[i][2]

		// Calculate ambient occlusion
		ao := m.calculateAO(
			int(float32(lx)+vx),
			int(float32(ly)+vy),
			int(float32(lz)+vz),
			face, c, getBlock,
		)
		aoFactor := 1.0 - ao*0.2

		// Position
		m.vertices = append(m.vertices, x+vx, y+vy, z+vz)
		// Normal
		m.vertices = append(m.vertices, normal[0], normal[1], normal[2])
		// Color with AO
		m.vertices = append(m.vertices, color[0]*aoFactor, color[1]*aoFactor, color[2]*aoFactor)
		// AO value
		m.vertices = append(m.vertices, ao)
		// Texture Coordinates
		m.vertices = append(m.vertices, faceUVs[i][0], faceUVs[i][1])
		// Material ID
		m.vertices = append(m.vertices, materialID)
	}

	// Two triangles per face
	m.indices = append(m.indices,
		baseIndex, baseIndex+1, baseIndex+2,
		baseIndex, baseIndex+2, baseIndex+3,
	)
}

// calculateAO calculates ambient occlusion for a vertex
func (m *Mesher) calculateAO(
	vx, vy, vz int,
	face string,
	c *Chunk,
	getBlock BlockGetter,
) float32 {
	// Simplified - count neighboring solid blocks
	count := 0
	offsets := [][3]int{
		{-1, 0, 0}, {1, 0, 0},
		{0, -1, 0}, {0, 1, 0},
		{0, 0, -1}, {0, 0, 1},
		{-1, -1, 0}, {1, 1, 0},
		{-1, 0, -1}, {1, 0, 1},
	}

	wx := int(c.CX)*Size + vx
	wz := int(c.CZ)*Size + vz

	for _, off := range offsets {
		neighbor := getBlock(wx+off[0], vy+off[1], wz+off[2])
		if neighbor != block.Air && !neighbor.IsTransparent() {
			count++
		}
	}

	ao := float32(count) / 4.0
	if ao > 1.0 {
		ao = 1.0
	}
	return ao
}

// resetBuffers clears the mesh buffers for reuse
func (m *Mesher) resetBuffers() {
	m.vertices = m.vertices[:0]
	m.indices = m.indices[:0]
}

// SharedMesher is a singleton mesher to avoid allocation overhead
var SharedMesher = NewMesher()
