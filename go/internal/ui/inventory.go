// Package ui provides inventory management
package ui

import (
	"voxelgame/internal/core/block"
)

// InventorySlot represents a slot in the inventory
type InventorySlot struct {
	BlockType block.Type
	Count     int
}

// Inventory manages player inventory
type Inventory struct {
	// Hotbar slots (9 slots)
	Hotbar [9]InventorySlot

	// Main inventory (27 slots)
	Main [27]InventorySlot

	// Currently selected hotbar slot
	SelectedIndex int

	// Is inventory open
	IsOpen bool
}

// NewInventory creates a new inventory with default blocks
func NewInventory() *Inventory {
	inv := &Inventory{
		SelectedIndex: 0,
	}

	// Initialize hotbar with common blocks
	defaults := []block.Type{
		block.Grass,
		block.Dirt,
		block.Stone,
		block.Wood,
		block.Glass,
		block.Brick,
		block.Sand,
		block.Water,
		block.OakLog,
	}

	for i, bt := range defaults {
		inv.Hotbar[i] = InventorySlot{BlockType: bt, Count: 64}
	}

	return inv
}

// GetSelectedBlock returns the currently selected block type
func (inv *Inventory) GetSelectedBlock() block.Type {
	if inv.SelectedIndex < 0 || inv.SelectedIndex >= len(inv.Hotbar) {
		return block.Air
	}
	return inv.Hotbar[inv.SelectedIndex].BlockType
}

// SelectSlot selects a hotbar slot by index (0-8)
func (inv *Inventory) SelectSlot(index int) {
	if index >= 0 && index < len(inv.Hotbar) {
		inv.SelectedIndex = index
	}
}

// ScrollSelection scrolls the hotbar selection
func (inv *Inventory) ScrollSelection(delta int) {
	inv.SelectedIndex += delta

	// Wrap around
	for inv.SelectedIndex < 0 {
		inv.SelectedIndex += len(inv.Hotbar)
	}
	inv.SelectedIndex %= len(inv.Hotbar)
}

// AddBlock adds a block to the inventory
func (inv *Inventory) AddBlock(bt block.Type, count int) bool {
	// First try to stack in hotbar
	for i := range inv.Hotbar {
		if inv.Hotbar[i].BlockType == bt && inv.Hotbar[i].Count < 64 {
			space := 64 - inv.Hotbar[i].Count
			add := count
			if add > space {
				add = space
			}
			inv.Hotbar[i].Count += add
			count -= add
			if count <= 0 {
				return true
			}
		}
	}

	// Then try main inventory
	for i := range inv.Main {
		if inv.Main[i].BlockType == bt && inv.Main[i].Count < 64 {
			space := 64 - inv.Main[i].Count
			add := count
			if add > space {
				add = space
			}
			inv.Main[i].Count += add
			count -= add
			if count <= 0 {
				return true
			}
		}
	}

	// Find empty slot in hotbar
	for i := range inv.Hotbar {
		if inv.Hotbar[i].Count == 0 {
			inv.Hotbar[i] = InventorySlot{BlockType: bt, Count: count}
			return true
		}
	}

	// Find empty slot in main
	for i := range inv.Main {
		if inv.Main[i].Count == 0 {
			inv.Main[i] = InventorySlot{BlockType: bt, Count: count}
			return true
		}
	}

	return false // Inventory full
}

// RemoveBlock removes a block from the selected slot
func (inv *Inventory) RemoveBlock() bool {
	if inv.Hotbar[inv.SelectedIndex].Count > 0 {
		inv.Hotbar[inv.SelectedIndex].Count--
		if inv.Hotbar[inv.SelectedIndex].Count <= 0 {
			inv.Hotbar[inv.SelectedIndex].BlockType = block.Air
		}
		return true
	}
	return false
}

// GetHotbarColors returns colors for hotbar display
func (inv *Inventory) GetHotbarColors() [][3]float32 {
	colors := make([][3]float32, len(inv.Hotbar))
	for i, slot := range inv.Hotbar {
		if slot.Count > 0 {
			colors[i] = slot.BlockType.GetColor()
		}
	}
	return colors
}

// Toggle opens/closes the inventory
func (inv *Inventory) Toggle() {
	inv.IsOpen = !inv.IsOpen
}
