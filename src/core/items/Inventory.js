/**
 * Player Inventory System
 */

import { itemRegistry } from "./ItemRegistry.js";
import { ItemCategory } from "./ItemTypes.js";

export class Inventory {
  constructor(size = 36) {
    this.size = size;
    this.slots = new Array(size).fill(null);
    this.selectedSlot = 0;
    this.hotbarSize = 9;
  }

  /**
   * Add item to inventory
   * @returns remaining count that couldn't be added
   */
  addItem(itemId, count = 1) {
    const itemDef = itemRegistry.get(itemId);
    if (!itemDef) return count;

    let remaining = count;
    const maxStack = itemDef.maxStack || 1;

    // First try to stack with existing items
    if (itemDef.stackable) {
      for (let i = 0; i < this.size && remaining > 0; i++) {
        const slot = this.slots[i];
        if (slot && slot.itemId === itemId && slot.count < maxStack) {
          const canAdd = Math.min(remaining, maxStack - slot.count);
          slot.count += canAdd;
          remaining -= canAdd;
        }
      }
    }

    // Then try to add to empty slots
    for (let i = 0; i < this.size && remaining > 0; i++) {
      if (!this.slots[i]) {
        const toAdd = Math.min(remaining, maxStack);
        this.slots[i] = {
          itemId,
          count: toAdd,
          durability: itemDef.durability || null,
        };
        remaining -= toAdd;
      }
    }

    return remaining;
  }

  /**
   * Remove item from inventory
   */
  removeItem(itemId, count = 1) {
    let remaining = count;

    for (let i = this.size - 1; i >= 0 && remaining > 0; i--) {
      const slot = this.slots[i];
      if (slot && slot.itemId === itemId) {
        const toRemove = Math.min(remaining, slot.count);
        slot.count -= toRemove;
        remaining -= toRemove;

        if (slot.count <= 0) {
          this.slots[i] = null;
        }
      }
    }

    return count - remaining; // Return how many were actually removed
  }

  /**
   * Get item in slot
   */
  getSlot(index) {
    return this.slots[index] || null;
  }

  /**
   * Set item in slot
   */
  setSlot(index, itemId, count = 1, durability = null) {
    if (index < 0 || index >= this.size) return false;

    if (itemId === null) {
      this.slots[index] = null;
    } else {
      const itemDef = itemRegistry.get(itemId);
      this.slots[index] = {
        itemId,
        count,
        durability: durability ?? itemDef?.durability ?? null,
      };
    }
    return true;
  }

  /**
   * Get currently selected item
   */
  getSelectedItem() {
    return this.slots[this.selectedSlot];
  }

  /**
   * Select hotbar slot
   */
  selectSlot(index) {
    if (index >= 0 && index < this.hotbarSize) {
      this.selectedSlot = index;
      return true;
    }
    return false;
  }

  /**
   * Count items of a type
   */
  countItem(itemId) {
    return this.slots.reduce((total, slot) => {
      if (slot && slot.itemId === itemId) {
        return total + slot.count;
      }
      return total;
    }, 0);
  }

  /**
   * Check if inventory has item
   */
  hasItem(itemId, count = 1) {
    return this.countItem(itemId) >= count;
  }

  /**
   * Get hotbar slots only
   */
  getHotbar() {
    return this.slots.slice(0, this.hotbarSize);
  }

  /**
   * Serialize inventory for saving
   */
  serialize() {
    return {
      size: this.size,
      slots: this.slots,
      selectedSlot: this.selectedSlot,
    };
  }

  /**
   * Load inventory from saved data
   */
  deserialize(data) {
    this.size = data.size || 36;
    this.slots = data.slots || new Array(this.size).fill(null);
    this.selectedSlot = data.selectedSlot || 0;
  }
}

export default Inventory;
