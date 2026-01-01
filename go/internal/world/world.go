// Package world provides the main world management
package world

import (
	"time"

	"voxelgame/internal/core/block"
	"voxelgame/internal/core/chunk"
	"voxelgame/internal/generation/entity"
	"voxelgame/internal/generation/terrain"
	"voxelgame/internal/render"

	"github.com/go-gl/mathgl/mgl32"
)

// World represents the voxel world
type World struct {
	// Seed for world generation
	Seed int64

	// Terrain generator
	TerrainGenerator *terrain.Generator

	// Chunk manager
	ChunkManager *chunk.Manager

	// Chunk renderer
	ChunkRenderer *render.ChunkRenderer

	// Chunk mesher
	Mesher *chunk.Mesher

	// Creature manager
	CreatureManager *CreatureManager

	// Player position for chunk loading
	playerX, playerY, playerZ float64

	// Stats
	chunksLoaded    int
	meshesGenerated int
	lastUpdateTime  time.Time
}

// NewWorld creates a new world with the given seed
func NewWorld(seed int64) *World {
	terrainGen := terrain.NewGenerator(seed)

	chunkConfig := chunk.DefaultManagerConfig()
	chunkConfig.RenderDistance = 3
	chunkConfig.MaxLoadedChunks = 100

	w := &World{
		Seed:             seed,
		TerrainGenerator: terrainGen,
		ChunkManager:     chunk.NewManager(chunkConfig, terrainGen),
		ChunkRenderer:    render.NewChunkRenderer(),
		Mesher:           chunk.NewMesher(),
		CreatureManager:  NewCreatureManager(seed),
		lastUpdateTime:   time.Now(),
	}

	// Set up callbacks
	w.ChunkManager.OnChunkLoaded = w.onChunkLoaded
	w.ChunkManager.OnChunkUnloaded = w.onChunkUnloaded

	return w
}

// Update updates the world based on player position
func (w *World) Update(playerX, playerY, playerZ float64) {
	w.playerX = playerX
	w.playerY = playerY
	w.playerZ = playerZ

	// Update chunks around player
	loadRequests := w.ChunkManager.UpdateAroundPlayer(playerX, playerZ)

	// Load a few chunks per frame
	maxLoadsPerFrame := 2
	for i := 0; i < len(loadRequests) && i < maxLoadsPerFrame; i++ {
		req := loadRequests[i]
		w.ChunkManager.LoadChunk(req.CX, req.CZ)
	}

	// Update dirty chunks
	dirtyChunks := w.ChunkManager.GetDirtyChunks()
	maxMeshesPerFrame := 2
	for i := 0; i < len(dirtyChunks) && i < maxMeshesPerFrame; i++ {
		w.regenerateMesh(dirtyChunks[i])
	}

	w.chunksLoaded = w.ChunkManager.LoadedCount()

	// Update creatures
	playerPos := mgl32.Vec3{float32(playerX), float32(playerY), float32(playerZ)}
	w.CreatureManager.Update(0.016, playerPos, w.GetBiomeAt, w.GetHeight)
}

// Render renders all visible chunks
func (w *World) Render() {
	w.ChunkRenderer.Draw()
}

// GetBlock returns the block at world coordinates
func (w *World) GetBlock(x, y, z int) block.Type {
	return w.ChunkManager.GetBlock(x, y, z)
}

// SetBlock sets a block at world coordinates
func (w *World) SetBlock(x, y, z int, t block.Type) bool {
	return w.ChunkManager.SetBlock(x, y, z, t)
}

// GetHeight returns terrain height at world coordinates
func (w *World) GetHeight(x, z int) int {
	return w.ChunkManager.GetHeight(x, z)
}

// GetSpawnPosition returns a suitable spawn position
func (w *World) GetSpawnPosition() (x, y, z float64) {
	spawnX, spawnZ := 0.0, 0.0

	// Load spawn chunk
	w.ChunkManager.LoadChunk(0, 0)

	// Get terrain height - find highest non-air block
	height := w.GetHeight(0, 0)

	// Spawn well above terrain to avoid being stuck
	spawnY := float64(height) + 10
	if spawnY < 50 {
		spawnY = 50 // Minimum spawn height
	}

	return spawnX, spawnY, spawnZ
}

// GetStats returns world statistics
func (w *World) GetStats() WorldStats {
	return WorldStats{
		ChunksLoaded:  w.chunksLoaded,
		MeshesLoaded:  w.ChunkRenderer.GetMeshCount(),
		CreatureCount: w.CreatureManager.GetCreatureCount(),
		Seed:          w.Seed,
	}
}

// WorldStats contains world statistics
type WorldStats struct {
	ChunksLoaded  int
	MeshesLoaded  int
	CreatureCount int
	Seed          int64
}

// Cleanup releases resources
func (w *World) Cleanup() {
	w.ChunkManager.Clear()
	w.ChunkRenderer.Cleanup()
	w.CreatureManager.Clear()
}

// Private methods

func (w *World) onChunkLoaded(c *chunk.Chunk) {
	w.regenerateMesh(c)
}

func (w *World) onChunkUnloaded(c *chunk.Chunk) {
	w.ChunkRenderer.RemoveChunk(c.ID())
}

func (w *World) regenerateMesh(c *chunk.Chunk) {
	// Block getter that uses world coordinates
	getBlock := func(wx, wy, wz int) block.Type {
		return w.ChunkManager.GetBlock(wx, wy, wz)
	}

	meshData := w.Mesher.GenerateMesh(c, getBlock)
	w.ChunkRenderer.UpdateChunk(c, meshData)
	w.meshesGenerated++
}

// GetBiomeAt returns the biome name at the given world coordinates
func (w *World) GetBiomeAt(x, z int) string {
	return w.TerrainGenerator.GetBiomeName(x, z)
}

// GetCreatures returns all creatures for rendering
func (w *World) GetCreatures() []*entity.Creature {
	return w.CreatureManager.GetCreatures()
}
