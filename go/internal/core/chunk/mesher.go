// Package chunk provides mesh generation for voxel chunks
package chunk

import (
	"voxelgame/internal/core/block"
)

// Vertex data layout for OpenGL
// Position (3) + Normal (3) + Color (3) + AO (1) + TexCoord (2) + MaterialID (1) + TextureLayerID (1) = 14 floats
const VertexSize = 14

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

				// Add custom details (foliage, grass blades)
				if blockDef.HasCustomMesh || blockType == block.Grass {
					m.addDetailedGeometry(
						x, y, z,
						worldX, y, worldZ,
						blockType, blockDef,
						c, getBlock,
					)
				}
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

// addDetailedGeometry adds custom geometry like cross-meshes or grass blades
func (m *Mesher) addDetailedGeometry(
	lx, ly, lz int,
	wx, wy, wz int,
	blockType block.Type,
	blockDef block.Definition,
	c *Chunk,
	getBlock BlockGetter,
) {
	// 1. Cross Mesh for Flowers/TallGrass (Foliage Material + Empty/Transparent block)
	// We use this if it's not a solid cube but has custom mesh
	if !blockDef.Solid && blockDef.HasCustomMesh {
		m.addCrossMesh(float32(lx), float32(ly), float32(lz), blockDef)
		return
	}

	// 2. Extra blades for Grass Block (Solid block with foliage on top)
	if blockType == block.Grass {
		// Only add blades if block above is air
		above := getBlock(wx, wy+1, wz)
		if above == block.Air || (block.GetDefinition(above).Transparent) {
			m.addGrassBlades(float32(lx), float32(ly), float32(lz), blockDef)
		}
	}
}

// addCrossMesh adds two intersecting quads for flowers/grass
func (m *Mesher) addCrossMesh(x, y, z float32, blockDef block.Definition) {
	color := blockDef.Color
	matID := float32(blockDef.Material)
	// We'll use the side texture for the cross pattern
	// In the future add texture support for this geometry

	// Quad 1: (0,0,0)-(1,0,1) diagonal? No, vertical diagonal.
	// (0.15, 0, 0.15) to (0.85, 1, 0.85)
	v1 := [4][3]float32{
		{0.15, 0, 0.15}, {0.85, 0, 0.85}, // Bottom
		{0.85, 1, 0.85}, {0.15, 1, 0.15}, // Top
	}
	// Quad 2
	v2 := [4][3]float32{
		{0.15, 0, 0.85}, {0.85, 0, 0.15}, // Bottom
		{0.85, 1, 0.15}, {0.15, 1, 0.85}, // Top
	}

	// Add both sides of diagonal 1
	m.addQuadCustom(v1[0], v1[1], v1[2], v1[3], x, y, z, color, matID, 0, 0, 1, 1)
	m.addQuadCustom(v1[1], v1[0], v1[3], v1[2], x, y, z, color, matID, 0, 0, 1, 1) // Flip

	// Add both sides of diagonal 2
	m.addQuadCustom(v2[0], v2[1], v2[2], v2[3], x, y, z, color, matID, 0, 0, 1, 1)
	m.addQuadCustom(v2[1], v2[0], v2[3], v2[2], x, y, z, color, matID, 0, 0, 1, 1) // Flip
}

// addGrassBlades adds small random blades on top of a block
func (m *Mesher) addGrassBlades(x, y, z float32, blockDef block.Definition) {
	// Add 4-5 random small triangles/quads
	// Pseudorandom based on position
	seed := int(x*31 + y*17 + z*23)
	count := 5 + (seed % 3)

	color := blockDef.Color
	matID := float32(block.MaterialFoliage) // Force foliage material for sway

	for i := 0; i < count; i++ {
		// Random offset
		r1 := float32((seed+i*13)%100) / 100.0
		r2 := float32((seed+i*77)%100) / 100.0

		ox := 0.1 + r1*0.8
		oz := 0.1 + r2*0.8

		h := 0.3 + r1*0.3   // Height 0.3-0.6
		w := 0.05 + r2*0.05 // Width

		// Small quad
		v1 := [3]float32{ox - w, 1.0, oz}
		v2 := [3]float32{ox + w, 1.0, oz}
		v3 := [3]float32{ox + w, 1.0 + h, oz + (r1-0.5)*0.2} // Lean slightly
		v4 := [3]float32{ox - w, 1.0 + h, oz + (r1-0.5)*0.2}

		m.addQuadCustom(v1, v2, v3, v4, x, y, z, color, matID, 0, 0, 1, 1)
	}
}

// addQuadCustom adds a custom quad with manual vertices
func (m *Mesher) addQuadCustom(
	p1, p2, p3, p4 [3]float32,
	x, y, z float32,
	color [3]float32,
	matID float32,
	u0, v0, u1, v1 float32,
) {
	baseIndex := uint32(len(m.vertices) / VertexSize)

	// Normal (simplified up)
	nx, ny, nz := float32(0), float32(1), float32(0)

	// Vertices
	verts := [][3]float32{p1, p2, p3, p4}
	uvs := [][2]float32{{u0, v0}, {u1, v0}, {u1, v1}, {u0, v1}}

	for i := 0; i < 4; i++ {
		// Pos
		m.vertices = append(m.vertices, x+verts[i][0], y+verts[i][1], z+verts[i][2])
		// Normal
		m.vertices = append(m.vertices, nx, ny, nz)
		// Color
		m.vertices = append(m.vertices, color[0], color[1], color[2])
		// AO (Full lit for details)
		m.vertices = append(m.vertices, 1.0)
		// UV
		m.vertices = append(m.vertices, uvs[i][0], uvs[i][1])
		// MatID
		m.vertices = append(m.vertices, matID)
		// TextureLayerID (use 0 for custom geometry, color-based)
		m.vertices = append(m.vertices, 0)
	}

	m.indices = append(m.indices,
		baseIndex, baseIndex+1, baseIndex+2,
		baseIndex, baseIndex+2, baseIndex+3,
	)
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
	// Special handling for water to prevent internal face flickering
	isWater := blockType == block.Water

	for i, face := range faceNames {
		offset := neighborOffsets[i]
		neighborType := getBlock(wx+offset[0], wy+offset[1], wz+offset[2])

		// Only add face if neighbor is air or transparent (and different type)
		neighborDef := block.GetDefinition(neighborType)

		shouldRender := false

		if isWater {
			// For water: only render face if neighbor is NOT water and NOT air
			// Exception: render top face if neighbor above is air (water surface)
			if face == "top" && neighborType == block.Air {
				shouldRender = true
			} else if neighborType != block.Water && neighborType != block.Air && !neighborDef.Transparent {
				// Side/bottom faces only when adjacent to solid blocks (for underwater viewing)
				shouldRender = true
			}
		} else {
			// Normal logic for other blocks
			if neighborType == block.Air ||
				(neighborDef.Transparent && neighborType != blockType) {
				shouldRender = true
			}
		}

		if shouldRender {
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

	// Select texture layer based on face
	var textureLayerID float32
	switch face {
	case "top":
		textureLayerID = float32(blockDef.TextureTop)
	case "bottom":
		textureLayerID = float32(blockDef.TextureBottom)
	default: // front, back, left, right
		textureLayerID = float32(blockDef.TextureSide)
	}

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
		// Texture Layer ID
		m.vertices = append(m.vertices, textureLayerID)
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
