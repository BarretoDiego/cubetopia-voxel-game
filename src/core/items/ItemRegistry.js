/**
 * Item Registry - manages all item definitions
 */

import { ItemTypes, ItemDefinitions, ItemCategory } from "./ItemTypes.js";

class ItemRegistry {
  constructor() {
    this.items = new Map();

    // Register all default items
    Object.entries(ItemDefinitions).forEach(([id, def]) => {
      this.items.set(parseInt(id), {
        id: parseInt(id),
        ...def,
      });
    });
  }

  /**
   * Get item definition by ID
   */
  get(id) {
    return this.items.get(id) || null;
  }

  /**
   * Check if item is a weapon
   */
  isWeapon(id) {
    const item = this.get(id);
    return item?.category === ItemCategory.WEAPON;
  }

  /**
   * Check if item is a tool
   */
  isTool(id) {
    const item = this.get(id);
    return item?.category === ItemCategory.TOOL;
  }

  /**
   * Check if item is stackable
   */
  isStackable(id) {
    const item = this.get(id);
    return item?.stackable !== false;
  }

  /**
   * Get max stack size for item
   */
  getMaxStack(id) {
    const item = this.get(id);
    return item?.maxStack || 1;
  }

  /**
   * Get all items of a category
   */
  getByCategory(category) {
    const result = [];
    this.items.forEach((item) => {
      if (item.category === category) {
        result.push(item);
      }
    });
    return result;
  }

  /**
   * Get all weapons
   */
  getWeapons() {
    return this.getByCategory(ItemCategory.WEAPON);
  }

  /**
   * Get all tools
   */
  getTools() {
    return this.getByCategory(ItemCategory.TOOL);
  }
}

export const itemRegistry = new ItemRegistry();
export default itemRegistry;
