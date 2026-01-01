// Package render provides mesh management for OpenGL
package render

import (
	"voxelgame/internal/core/chunk"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// ChunkMesh manages OpenGL buffers for a chunk mesh
type ChunkMesh struct {
	VAO         uint32
	VBO         uint32
	EBO         uint32
	VertexCount int32
	IndexCount  int32
}

// NewChunkMesh creates OpenGL buffers from mesh data
func NewChunkMesh(data *chunk.MeshData) *ChunkMesh {
	if data == nil || data.VertexCount == 0 {
		return nil
	}

	mesh := &ChunkMesh{
		VertexCount: int32(data.VertexCount),
		IndexCount:  int32(data.IndexCount),
	}

	// Create VAO
	gl.GenVertexArrays(1, &mesh.VAO)
	gl.BindVertexArray(mesh.VAO)

	// Create VBO
	gl.GenBuffers(1, &mesh.VBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, mesh.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(data.Vertices)*4, gl.Ptr(data.Vertices), gl.STATIC_DRAW)

	// Create EBO
	gl.GenBuffers(1, &mesh.EBO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, mesh.EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(data.Indices)*4, gl.Ptr(data.Indices), gl.STATIC_DRAW)

	// Vertex layout: Position (3) + Normal (3) + Color (3) + AO (1) = 10 floats
	stride := int32(chunk.VertexSize * 4)

	// Position attribute (location 0)
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, stride, 0)
	gl.EnableVertexAttribArray(0)

	// Normal attribute (location 1)
	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, stride, 3*4)
	gl.EnableVertexAttribArray(1)

	// Color attribute (location 2)
	gl.VertexAttribPointerWithOffset(2, 3, gl.FLOAT, false, stride, 6*4)
	gl.EnableVertexAttribArray(2)

	// AO attribute (location 3)
	gl.VertexAttribPointerWithOffset(3, 1, gl.FLOAT, false, stride, 9*4)
	gl.EnableVertexAttribArray(3)

	// Unbind
	gl.BindVertexArray(0)

	return mesh
}

// Draw renders the mesh
func (m *ChunkMesh) Draw() {
	if m == nil || m.VAO == 0 {
		return
	}

	gl.BindVertexArray(m.VAO)
	gl.DrawElements(gl.TRIANGLES, m.IndexCount, gl.UNSIGNED_INT, nil)
	gl.BindVertexArray(0)
}

// Delete cleans up OpenGL resources
func (m *ChunkMesh) Delete() {
	if m == nil {
		return
	}

	if m.VAO != 0 {
		gl.DeleteVertexArrays(1, &m.VAO)
		m.VAO = 0
	}
	if m.VBO != 0 {
		gl.DeleteBuffers(1, &m.VBO)
		m.VBO = 0
	}
	if m.EBO != 0 {
		gl.DeleteBuffers(1, &m.EBO)
		m.EBO = 0
	}
}

// ChunkRenderer manages rendering of all chunk meshes
type ChunkRenderer struct {
	meshes map[string]*ChunkMesh
}

// NewChunkRenderer creates a new chunk renderer
func NewChunkRenderer() *ChunkRenderer {
	return &ChunkRenderer{
		meshes: make(map[string]*ChunkMesh),
	}
}

// UpdateChunk creates or updates mesh for a chunk
func (r *ChunkRenderer) UpdateChunk(c *chunk.Chunk, data *chunk.MeshData) {
	id := c.ID()

	// Delete old mesh if exists
	if old, ok := r.meshes[id]; ok {
		old.Delete()
		delete(r.meshes, id)
	}

	// Create new mesh
	if data != nil && data.VertexCount > 0 {
		r.meshes[id] = NewChunkMesh(data)
	}

	c.IsDirty = false
}

// RemoveChunk removes a chunk mesh
func (r *ChunkRenderer) RemoveChunk(id string) {
	if mesh, ok := r.meshes[id]; ok {
		mesh.Delete()
		delete(r.meshes, id)
	}
}

// Draw renders all chunk meshes
func (r *ChunkRenderer) Draw() {
	for _, mesh := range r.meshes {
		mesh.Draw()
	}
}

// GetMeshCount returns number of loaded meshes
func (r *ChunkRenderer) GetMeshCount() int {
	return len(r.meshes)
}

// Cleanup removes all meshes
func (r *ChunkRenderer) Cleanup() {
	for id, mesh := range r.meshes {
		mesh.Delete()
		delete(r.meshes, id)
	}
}
