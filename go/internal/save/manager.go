// Package save provides save/load functionality
package save

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SaveData contains all game state to be saved
type SaveData struct {
	Version   string     `json:"version"`
	Timestamp int64      `json:"timestamp"`
	Player    PlayerSave `json:"player"`
	World     WorldSave  `json:"world"`
}

// PlayerSave contains player state
type PlayerSave struct {
	PositionX float32 `json:"x"`
	PositionY float32 `json:"y"`
	PositionZ float32 `json:"z"`
	Yaw       float32 `json:"yaw"`
	Pitch     float32 `json:"pitch"`
}

// WorldSave contains world state
type WorldSave struct {
	Seed           int64                   `json:"seed"`
	ModifiedChunks map[string]ChunkModSave `json:"modifiedChunks"`
}

// ChunkModSave contains modifications to a chunk
type ChunkModSave struct {
	CX            int            `json:"cx"`
	CZ            int            `json:"cz"`
	Modifications []BlockModSave `json:"modifications"`
}

// BlockModSave contains a single block modification
type BlockModSave struct {
	X    int   `json:"x"`
	Y    int   `json:"y"`
	Z    int   `json:"z"`
	Type uint8 `json:"type"`
}

// Manager handles save/load operations
type Manager struct {
	saveDir string
}

// NewManager creates a new save manager
func NewManager() *Manager {
	// Use user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	saveDir := filepath.Join(home, ".voxelgame", "saves")

	// Create save directory if it doesn't exist
	_ = os.MkdirAll(saveDir, 0755)

	return &Manager{
		saveDir: saveDir,
	}
}

// Save saves the game state
func (m *Manager) Save(saveName string, data SaveData) error {
	data.Version = "1.0"
	data.Timestamp = time.Now().Unix()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal save data: %w", err)
	}

	path := filepath.Join(m.saveDir, saveName+".json")
	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write save file: %w", err)
	}

	fmt.Printf("[SaveManager] Saved game to %s\n", path)
	return nil
}

// Load loads the game state
func (m *Manager) Load(saveName string) (*SaveData, error) {
	path := filepath.Join(m.saveDir, saveName+".json")

	jsonData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read save file: %w", err)
	}

	var data SaveData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to parse save data: %w", err)
	}

	fmt.Printf("[SaveManager] Loaded game from %s\n", path)
	return &data, nil
}

// ListSaves returns a list of available saves
func (m *Manager) ListSaves() ([]SaveInfo, error) {
	entries, err := os.ReadDir(m.saveDir)
	if err != nil {
		return nil, err
	}

	var saves []SaveInfo
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		name := entry.Name()[:len(entry.Name())-5] // Remove .json

		info, err := entry.Info()
		if err != nil {
			continue
		}

		saves = append(saves, SaveInfo{
			Name:      name,
			Timestamp: info.ModTime().Unix(),
		})
	}

	return saves, nil
}

// DeleteSave deletes a save file
func (m *Manager) DeleteSave(saveName string) error {
	path := filepath.Join(m.saveDir, saveName+".json")
	return os.Remove(path)
}

// Exists checks if a save exists
func (m *Manager) Exists(saveName string) bool {
	path := filepath.Join(m.saveDir, saveName+".json")
	_, err := os.Stat(path)
	return err == nil
}

// SaveInfo contains information about a save
type SaveInfo struct {
	Name      string `json:"name"`
	Timestamp int64  `json:"timestamp"`
}

// QuickSave is a convenience function for quick saving
func (m *Manager) QuickSave(player PlayerSave, worldSeed int64) error {
	return m.Save("quicksave", SaveData{
		Player: player,
		World: WorldSave{
			Seed:           worldSeed,
			ModifiedChunks: make(map[string]ChunkModSave),
		},
	})
}

// QuickLoad is a convenience function for quick loading
func (m *Manager) QuickLoad() (*SaveData, error) {
	return m.Load("quicksave")
}
