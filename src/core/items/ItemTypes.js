/**
 * Item Types and Definitions
 * Defines all items, weapons, and tools in the game
 */

// Item Categories
export const ItemCategory = {
  BLOCK: "block",
  WEAPON: "weapon",
  TOOL: "tool",
  CONSUMABLE: "consumable",
  MATERIAL: "material",
};

// Item Types
export const ItemTypes = {
  // Weapons
  WOODEN_SWORD: 100,
  STONE_SWORD: 101,
  IRON_SWORD: 102,

  // Tools
  WOODEN_PICKAXE: 200,
  STONE_PICKAXE: 201,
  IRON_PICKAXE: 202,
  WOODEN_AXE: 210,
  STONE_AXE: 211,

  // Consumables
  APPLE: 300,
  BREAD: 301,

  // Materials
  STICK: 400,
  IRON_INGOT: 401,
  DIAMOND: 402,
};

// Item Definitions
export const ItemDefinitions = {
  // Weapons
  [ItemTypes.WOODEN_SWORD]: {
    name: "Espada de Madeira",
    category: ItemCategory.WEAPON,
    damage: 4,
    durability: 60,
    attackSpeed: 1.2,
    color: "#8B4513",
    stackable: false,
  },
  [ItemTypes.STONE_SWORD]: {
    name: "Espada de Pedra",
    category: ItemCategory.WEAPON,
    damage: 5,
    durability: 132,
    attackSpeed: 1.2,
    color: "#808080",
    stackable: false,
  },
  [ItemTypes.IRON_SWORD]: {
    name: "Espada de Ferro",
    category: ItemCategory.WEAPON,
    damage: 6,
    durability: 250,
    attackSpeed: 1.2,
    color: "#C0C0C0",
    stackable: false,
  },

  // Pickaxes
  [ItemTypes.WOODEN_PICKAXE]: {
    name: "Picareta de Madeira",
    category: ItemCategory.TOOL,
    miningSpeed: 1.5,
    durability: 60,
    color: "#8B4513",
    stackable: false,
  },
  [ItemTypes.STONE_PICKAXE]: {
    name: "Picareta de Pedra",
    category: ItemCategory.TOOL,
    miningSpeed: 2.0,
    durability: 132,
    color: "#808080",
    stackable: false,
  },
  [ItemTypes.IRON_PICKAXE]: {
    name: "Picareta de Ferro",
    category: ItemCategory.TOOL,
    miningSpeed: 3.0,
    durability: 250,
    color: "#C0C0C0",
    stackable: false,
  },

  // Axes
  [ItemTypes.WOODEN_AXE]: {
    name: "Machado de Madeira",
    category: ItemCategory.TOOL,
    damage: 3,
    miningSpeed: 1.5,
    durability: 60,
    color: "#8B4513",
    stackable: false,
  },
  [ItemTypes.STONE_AXE]: {
    name: "Machado de Pedra",
    category: ItemCategory.TOOL,
    damage: 4,
    miningSpeed: 2.0,
    durability: 132,
    color: "#808080",
    stackable: false,
  },

  // Consumables
  [ItemTypes.APPLE]: {
    name: "Maçã",
    category: ItemCategory.CONSUMABLE,
    healing: 4,
    color: "#FF0000",
    stackable: true,
    maxStack: 64,
  },
  [ItemTypes.BREAD]: {
    name: "Pão",
    category: ItemCategory.CONSUMABLE,
    healing: 5,
    color: "#DEB887",
    stackable: true,
    maxStack: 64,
  },

  // Materials
  [ItemTypes.STICK]: {
    name: "Graveto",
    category: ItemCategory.MATERIAL,
    color: "#8B4513",
    stackable: true,
    maxStack: 64,
  },
  [ItemTypes.IRON_INGOT]: {
    name: "Barra de Ferro",
    category: ItemCategory.MATERIAL,
    color: "#C0C0C0",
    stackable: true,
    maxStack: 64,
  },
  [ItemTypes.DIAMOND]: {
    name: "Diamante",
    category: ItemCategory.MATERIAL,
    color: "#00FFFF",
    stackable: true,
    maxStack: 64,
  },
};

export default { ItemCategory, ItemTypes, ItemDefinitions };
