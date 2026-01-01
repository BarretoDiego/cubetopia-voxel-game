// Package world provides the main world management
package world

import (
	"time"

	"fmt"
	"voxelgame/internal/core/block"
	"voxelgame/internal/core/chunk"
	"voxelgame/internal/generation/entity"
	"voxelgame/internal/generation/terrain"
	"voxelgame/internal/render"
	"voxelgame/internal/save"

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

	// Save manager
	SaveManager *save.Manager

	// Player position for chunk loading
	playerX, playerY, playerZ float64

	// Stats
	chunksLoaded    int
	meshesGenerated int
	lastUpdateTime  time.Time
	// Time of Day system
	TimeOfDay *TimeOfDay
}

// NewWorld creates a new world with the given seed
func NewWorld(seed int64) *World {
	terrainGen := terrain.NewGenerator(seed)

	chunkConfig := chunk.DefaultManagerConfig()
	chunkConfig.RenderDistance = 10 // Default render distance
	chunkConfig.MaxLoadedChunks = 200

	w := &World{
		Seed:             seed,
		TerrainGenerator: terrainGen,
		ChunkManager:     chunk.NewManager(chunkConfig, terrainGen),
		ChunkRenderer:    render.NewChunkRenderer(),
		Mesher:           chunk.NewMesher(),
		CreatureManager:  NewCreatureManager(seed),
		SaveManager:      save.NewManager(),
		lastUpdateTime:   time.Now(),
		TimeOfDay:        NewTimeOfDay(),
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

	// Calculate delta time
	dt := float32(time.Since(w.lastUpdateTime).Seconds())
	w.lastUpdateTime = time.Now()
	if dt > 0.1 {
		dt = 0.1 // Cap max dt to prevent huge jumps
	}

	// Update time of day
	w.TimeOfDay.Update(dt)

	// Update chunks around player
	loadRequests := w.ChunkManager.UpdateAroundPlayer(playerX, playerZ)

	// Load more chunks per frame to reduce gaps
	maxLoadsPerFrame := 4
	for i := 0; i < len(loadRequests) && i < maxLoadsPerFrame; i++ {
		req := loadRequests[i]
		w.ChunkManager.LoadChunk(req.CX, req.CZ)
	}

	// Update dirty chunks
	dirtyChunks := w.ChunkManager.GetDirtyChunks()
	// Process more meshes per frame to handle updates faster
	maxMeshesPerFrame := 32
	for i := 0; i < len(dirtyChunks) && i < maxMeshesPerFrame; i++ {
		w.regenerateMesh(dirtyChunks[i])
	}

	w.chunksLoaded = w.ChunkManager.LoadedCount()

	// Update creatures
	playerPos := mgl32.Vec3{float32(playerX), float32(playerY), float32(playerZ)}
	w.CreatureManager.Update(0.016, playerPos, w.GetBiomeAt, w.GetHeight)
}

// ApplySettings applies settings to the world
func (w *World) ApplySettings(dayDuration, nightBrightness float32, terrainConfig terrain.GeneratorConfig) {
	if w.TimeOfDay != nil {
		w.TimeOfDay.DayDurationSeconds = dayDuration
		w.TimeOfDay.NightBrightness = nightBrightness
	}
	if w.TerrainGenerator != nil {
		w.TerrainGenerator.SetConfig(terrainConfig)
	}
}

// Render renders all visible chunks
// Render renders all visible chunks
func (w *World) Render() {
	if w.ChunkRenderer == nil {
		return
	}
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
	// Auto-save on exit
	if err := w.Save("autosave"); err != nil {
		fmt.Printf("Failed to autosave: %v\n", err)
	}

	w.ChunkManager.Clear()
	w.ChunkRenderer.Cleanup()
	w.CreatureManager.Clear()
}

// Save saves the world state
func (w *World) Save(saveName string) error {
	// Get all modifications from chunk manager
	modifications := w.ChunkManager.GetAllModifications()

	// Convert to save format
	saveMods := make(map[string]save.ChunkModSave)
	for id, mods := range modifications {
		var saveBlockMods []save.BlockModSave
		for _, m := range mods {
			saveBlockMods = append(saveBlockMods, save.BlockModSave{
				X:    m.X,
				Y:    m.Y,
				Z:    m.Z,
				Type: uint8(m.Type),
			})
		}

		// Parse chunk ID to get CX, CZ
		var cx, cz int
		fmt.Sscanf(id, "%d,%d", &cx, &cz)

		saveMods[id] = save.ChunkModSave{
			CX:            cx,
			CZ:            cz,
			Modifications: saveBlockMods,
		}
	}

	playerSave := save.PlayerSave{
		PositionX: float32(w.playerX),
		PositionY: float32(w.playerY),
		PositionZ: float32(w.playerZ),
		// Yaw and Pitch would need to be passed in or stored in World if we want to save them
	}

	worldSave := save.WorldSave{
		Seed:           w.Seed,
		ModifiedChunks: saveMods,
	}

	return w.SaveManager.Save(saveName, save.SaveData{
		Player: playerSave,
		World:  worldSave,
	})
}

// Load loads the world state
func (w *World) Load(saveName string) error {
	data, err := w.SaveManager.Load(saveName)
	if err != nil {
		return err
	}

	w.Seed = data.World.Seed
	// Re-initialize generator with saved seed
	w.TerrainGenerator = terrain.NewGenerator(w.Seed)
	// Re-create manager with new generator (keeps config)
	config := chunk.DefaultManagerConfig()
	config.RenderDistance = 10
	config.MaxLoadedChunks = 200
	w.ChunkManager = chunk.NewManager(config, w.TerrainGenerator)

	// Convert modifications back to chunk manager format
	chunkMods := make(map[string][]chunk.BlockModificationWorld)
	for id, modSave := range data.World.ModifiedChunks {
		var blockMods []chunk.BlockModificationWorld
		for _, m := range modSave.Modifications {
			blockMods = append(blockMods, chunk.BlockModificationWorld{
				X:    m.X,
				Y:    m.Y,
				Z:    m.Z,
				Type: block.Type(m.Type),
			})
		}
		chunkMods[id] = blockMods
	}

	w.ChunkManager.SetModifications(chunkMods)

	// Set player position
	w.playerX = float64(data.Player.PositionX)
	w.playerY = float64(data.Player.PositionY)
	w.playerZ = float64(data.Player.PositionZ)

	// Setup callbacks again since we recreated the manager
	w.ChunkManager.OnChunkLoaded = w.onChunkLoaded
	w.ChunkManager.OnChunkUnloaded = w.onChunkUnloaded

	return nil
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
