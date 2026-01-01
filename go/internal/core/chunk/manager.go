// Package chunk provides chunk management with loading/unloading
package chunk

import (
	"fmt"
	"math"
	"sync"

	"voxelgame/internal/core/block"
)

// Manager handles chunk loading, unloading, and caching
type Manager struct {
	// Active chunks
	chunks map[string]*Chunk
	mu     sync.RWMutex

	// LRU cache for unloaded chunks
	cache      map[string]*Chunk
	cacheOrder []string
	cacheMu    sync.Mutex

	// Configuration
	maxLoadedChunks int
	maxCachedChunks int
	renderDistance  int

	// Generator interface for chunk generation
	generator ChunkGenerator

	// Event callbacks
	OnChunkLoaded   func(*Chunk)
	OnChunkUnloaded func(*Chunk)
}

// ChunkGenerator interface for terrain generation
type ChunkGenerator interface {
	GenerateChunk(c *Chunk)
}

// ManagerConfig holds configuration for the chunk manager
type ManagerConfig struct {
	MaxLoadedChunks int
	MaxCachedChunks int
	RenderDistance  int
}

// DefaultManagerConfig returns default manager configuration
func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		MaxLoadedChunks: 100,
		MaxCachedChunks: 20,
		RenderDistance:  3,
	}
}

// NewManager creates a new chunk manager
func NewManager(config ManagerConfig, generator ChunkGenerator) *Manager {
	return &Manager{
		chunks:          make(map[string]*Chunk),
		cache:           make(map[string]*Chunk),
		cacheOrder:      make([]string, 0, config.MaxCachedChunks),
		maxLoadedChunks: config.MaxLoadedChunks,
		maxCachedChunks: config.MaxCachedChunks,
		renderDistance:  config.RenderDistance,
		generator:       generator,
	}
}

// GetChunk returns a chunk if it's loaded, nil otherwise
func (m *Manager) GetChunk(cx, cz int) *Chunk {
	id := m.chunkID(cx, cz)

	m.mu.RLock()
	chunk, exists := m.chunks[id]
	m.mu.RUnlock()

	if exists {
		return chunk
	}

	// Check cache
	m.cacheMu.Lock()
	if cached, ok := m.cache[id]; ok {
		delete(m.cache, id)
		m.removeFromCacheOrder(id)
		m.cacheMu.Unlock()

		m.mu.Lock()
		m.chunks[id] = cached
		m.mu.Unlock()

		return cached
	}
	m.cacheMu.Unlock()

	return nil
}

// LoadChunk loads or generates a chunk at the given coordinates
func (m *Manager) LoadChunk(cx, cz int) *Chunk {
	id := m.chunkID(cx, cz)

	// Check if already loaded
	m.mu.RLock()
	if chunk, exists := m.chunks[id]; exists {
		m.mu.RUnlock()
		return chunk
	}
	m.mu.RUnlock()

	// Enforce chunk limit
	if m.LoadedCount() >= m.maxLoadedChunks {
		m.enforceChunkLimit()
		if m.LoadedCount() >= m.maxLoadedChunks {
			fmt.Printf("[ChunkManager] Chunk limit reached, cannot load %s\n", id)
			return nil
		}
	}

	// Check cache first
	m.cacheMu.Lock()
	if cached, ok := m.cache[id]; ok {
		delete(m.cache, id)
		m.removeFromCacheOrder(id)
		m.cacheMu.Unlock()

		m.mu.Lock()
		m.chunks[id] = cached
		m.mu.Unlock()

		return cached
	}
	m.cacheMu.Unlock()

	// Create and generate new chunk
	chunk := New(int32(cx), int32(cz))

	if m.generator != nil {
		m.generator.GenerateChunk(chunk)
	}

	chunk.IsGenerated = true

	m.mu.Lock()
	m.chunks[id] = chunk
	m.mu.Unlock()

	if m.OnChunkLoaded != nil {
		m.OnChunkLoaded(chunk)
	}

	return chunk
}

// UnloadChunk moves a chunk to the cache
func (m *Manager) UnloadChunk(cx, cz int) {
	id := m.chunkID(cx, cz)

	m.mu.Lock()
	chunk, exists := m.chunks[id]
	if exists {
		delete(m.chunks, id)
	}
	m.mu.Unlock()

	if !exists {
		return
	}

	// Add to cache
	m.cacheMu.Lock()
	// Evict oldest if cache is full
	for len(m.cache) >= m.maxCachedChunks && len(m.cacheOrder) > 0 {
		oldestID := m.cacheOrder[0]
		m.cacheOrder = m.cacheOrder[1:]
		if oldChunk, ok := m.cache[oldestID]; ok {
			oldChunk.Dispose()
			delete(m.cache, oldestID)
		}
	}
	m.cache[id] = chunk
	m.cacheOrder = append(m.cacheOrder, id)
	m.cacheMu.Unlock()

	if m.OnChunkUnloaded != nil {
		m.OnChunkUnloaded(chunk)
	}
}

// UpdateAroundPlayer loads/unloads chunks based on player position
func (m *Manager) UpdateAroundPlayer(playerX, playerZ float64) []ChunkLoadRequest {
	playerCX := int(math.Floor(playerX / float64(Size)))
	playerCZ := int(math.Floor(playerZ / float64(Size)))

	toKeep := make(map[string]bool)
	var toLoad []ChunkLoadRequest

	// Determine chunks that should be loaded
	for dx := -m.renderDistance; dx <= m.renderDistance; dx++ {
		for dz := -m.renderDistance; dz <= m.renderDistance; dz++ {
			cx := playerCX + dx
			cz := playerCZ + dz
			id := m.chunkID(cx, cz)

			toKeep[id] = true

			if m.GetChunk(cx, cz) == nil {
				dist := abs(dx) + abs(dz)
				toLoad = append(toLoad, ChunkLoadRequest{CX: cx, CZ: cz, Distance: dist})
			}
		}
	}

	// Unload chunks that are too far
	m.mu.RLock()
	var toUnload []struct{ cx, cz int }
	for id, chunk := range m.chunks {
		if !toKeep[id] {
			toUnload = append(toUnload, struct{ cx, cz int }{int(chunk.CX), int(chunk.CZ)})
		}
	}
	m.mu.RUnlock()

	for _, u := range toUnload {
		m.UnloadChunk(u.cx, u.cz)
	}

	// Sort by distance (closest first)
	sortByDistance(toLoad)

	return toLoad
}

// ChunkLoadRequest represents a request to load a chunk
type ChunkLoadRequest struct {
	CX, CZ   int
	Distance int
}

// GetBlock returns the block at world coordinates
func (m *Manager) GetBlock(wx, wy, wz int) block.Type {
	cx := int(math.Floor(float64(wx) / float64(Size)))
	cz := int(math.Floor(float64(wz) / float64(Size)))

	chunk := m.GetChunk(cx, cz)
	if chunk == nil {
		return block.Air
	}

	lx := mod(wx, Size)
	lz := mod(wz, Size)

	return chunk.GetBlock(lx, wy, lz)
}

// SetBlock sets the block at world coordinates
func (m *Manager) SetBlock(wx, wy, wz int, t block.Type) bool {
	cx := int(math.Floor(float64(wx) / float64(Size)))
	cz := int(math.Floor(float64(wz) / float64(Size)))

	chunk := m.GetChunk(cx, cz)
	if chunk == nil {
		return false
	}

	lx := mod(wx, Size)
	lz := mod(wz, Size)

	result := chunk.SetBlock(lx, wy, lz, t)

	// Mark neighboring chunks dirty if block is on edge
	if result {
		if lx == 0 {
			m.markDirty(cx-1, cz)
		}
		if lx == Size-1 {
			m.markDirty(cx+1, cz)
		}
		if lz == 0 {
			m.markDirty(cx, cz-1)
		}
		if lz == Size-1 {
			m.markDirty(cx, cz+1)
		}
	}

	return result
}

// GetHeight returns terrain height at world coordinates
func (m *Manager) GetHeight(wx, wz int) int {
	cx := int(math.Floor(float64(wx) / float64(Size)))
	cz := int(math.Floor(float64(wz) / float64(Size)))

	chunk := m.GetChunk(cx, cz)
	if chunk == nil {
		return 0
	}

	lx := mod(wx, Size)
	lz := mod(wz, Size)

	return chunk.GetHeight(lx, lz)
}

// GetLoadedChunks returns all currently loaded chunks
func (m *Manager) GetLoadedChunks() []*Chunk {
	m.mu.RLock()
	defer m.mu.RUnlock()

	chunks := make([]*Chunk, 0, len(m.chunks))
	for _, c := range m.chunks {
		chunks = append(chunks, c)
	}
	return chunks
}

// GetDirtyChunks returns chunks that need mesh rebuilding
func (m *Manager) GetDirtyChunks() []*Chunk {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dirty := make([]*Chunk, 0)
	for _, c := range m.chunks {
		if c.IsDirty {
			dirty = append(dirty, c)
		}
	}
	return dirty
}

// LoadedCount returns number of loaded chunks
func (m *Manager) LoadedCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.chunks)
}

// Clear unloads all chunks
func (m *Manager) Clear() {
	m.mu.Lock()
	for _, c := range m.chunks {
		c.Dispose()
	}
	m.chunks = make(map[string]*Chunk)
	m.mu.Unlock()

	m.cacheMu.Lock()
	for _, c := range m.cache {
		c.Dispose()
	}
	m.cache = make(map[string]*Chunk)
	m.cacheOrder = m.cacheOrder[:0]
	m.cacheMu.Unlock()
}

// Helper methods

func (m *Manager) chunkID(cx, cz int) string {
	return fmt.Sprintf("%d,%d", cx, cz)
}

func (m *Manager) enforceChunkLimit() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.chunks) <= m.maxLoadedChunks {
		return
	}

	fmt.Printf("[ChunkManager] Over limit! %d/%d chunks. Forcing cleanup...\n",
		len(m.chunks), m.maxLoadedChunks)

	// Sort by distance from origin
	type chunkDist struct {
		id    string
		chunk *Chunk
		dist  int
	}

	var chunks []chunkDist
	for id, c := range m.chunks {
		dist := abs(int(c.CX)) + abs(int(c.CZ))
		chunks = append(chunks, chunkDist{id, c, dist})
	}

	// Sort farthest first
	for i := 0; i < len(chunks)-1; i++ {
		for j := i + 1; j < len(chunks); j++ {
			if chunks[i].dist < chunks[j].dist {
				chunks[i], chunks[j] = chunks[j], chunks[i]
			}
		}
	}

	// Unload excess
	excess := len(m.chunks) - m.maxLoadedChunks
	for i := 0; i < excess && i < len(chunks); i++ {
		chunks[i].chunk.Dispose()
		delete(m.chunks, chunks[i].id)
	}
}

func (m *Manager) markDirty(cx, cz int) {
	if chunk := m.GetChunk(cx, cz); chunk != nil {
		chunk.IsDirty = true
	}
}

func (m *Manager) removeFromCacheOrder(id string) {
	for i, cid := range m.cacheOrder {
		if cid == id {
			m.cacheOrder = append(m.cacheOrder[:i], m.cacheOrder[i+1:]...)
			return
		}
	}
}

func mod(n, m int) int {
	return ((n % m) + m) % m
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func sortByDistance(requests []ChunkLoadRequest) {
	for i := 0; i < len(requests)-1; i++ {
		for j := i + 1; j < len(requests); j++ {
			if requests[i].Distance > requests[j].Distance {
				requests[i], requests[j] = requests[j], requests[i]
			}
		}
	}
}
