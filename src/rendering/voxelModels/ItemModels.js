/**
 * Item Models - High-resolution voxel models for weapons and tools
 */

import {
  createEmptyVoxelArray,
  fillBox,
  setVoxel,
  createVoxelGeometry,
} from "./VoxelModelRenderer.js";

/**
 * Create a Sword model
 */
export function createSwordModel(
  bladeColor = "#C0C0C0",
  handleColor = "#8B4513",
  size = 1
) {
  const voxels = createEmptyVoxelArray(3, 16, 1);
  const highlight = lightenColor(bladeColor, 0.3);

  // Handle
  fillBox(voxels, 1, 0, 0, 1, 3, 0, handleColor);

  // Guard
  fillBox(voxels, 0, 4, 0, 2, 4, 0, bladeColor);

  // Blade
  fillBox(voxels, 1, 5, 0, 1, 14, 0, bladeColor);

  // Blade tip
  setVoxel(voxels, 1, 15, 0, bladeColor);

  // Blade highlight (edge)
  for (let y = 5; y < 15; y++) {
    setVoxel(voxels, 0, y, 0, highlight);
  }

  return createVoxelGeometry(voxels, size);
}

/**
 * Create a Pickaxe model
 */
export function createPickaxeModel(
  headColor = "#808080",
  handleColor = "#8B4513",
  size = 1
) {
  const voxels = createEmptyVoxelArray(7, 16, 2);

  // Handle
  fillBox(voxels, 3, 0, 0, 3, 10, 1, handleColor);

  // Head
  fillBox(voxels, 0, 11, 0, 6, 13, 1, headColor);

  // Pick points
  setVoxel(voxels, 0, 14, 0, headColor);
  setVoxel(voxels, 0, 14, 1, headColor);
  setVoxel(voxels, 6, 14, 0, headColor);
  setVoxel(voxels, 6, 14, 1, headColor);

  return createVoxelGeometry(voxels, size);
}

/**
 * Create an Axe model
 */
export function createAxeModel(
  headColor = "#808080",
  handleColor = "#8B4513",
  size = 1
) {
  const voxels = createEmptyVoxelArray(5, 16, 2);

  // Handle
  fillBox(voxels, 2, 0, 0, 2, 10, 1, handleColor);

  // Axe head
  fillBox(voxels, 2, 11, 0, 4, 15, 1, headColor);

  // Axe blade edge
  setVoxel(voxels, 4, 15, 0, headColor);
  setVoxel(voxels, 4, 15, 1, headColor);
  setVoxel(voxels, 4, 10, 0, headColor);
  setVoxel(voxels, 4, 10, 1, headColor);

  return createVoxelGeometry(voxels, size);
}

/**
 * Create a Shovel model
 */
export function createShovelModel(
  headColor = "#808080",
  handleColor = "#8B4513",
  size = 1
) {
  const voxels = createEmptyVoxelArray(3, 16, 2);

  // Handle
  fillBox(voxels, 1, 0, 0, 1, 10, 1, handleColor);

  // Shovel head
  fillBox(voxels, 0, 11, 0, 2, 15, 1, headColor);

  // Rounded tip
  setVoxel(voxels, 0, 15, 0, null);
  setVoxel(voxels, 0, 15, 1, null);
  setVoxel(voxels, 2, 15, 0, null);
  setVoxel(voxels, 2, 15, 1, null);

  return createVoxelGeometry(voxels, size);
}

/**
 * Create an Apple model
 */
export function createAppleModel(size = 1) {
  const voxels = createEmptyVoxelArray(8, 10, 8);
  const red = "#FF0000";
  const darkRed = "#8B0000";
  const stem = "#8B4513";
  const leaf = "#228B22";

  // Main body - spherical apple
  for (let x = 1; x < 7; x++) {
    for (let y = 1; y < 7; y++) {
      for (let z = 1; z < 7; z++) {
        const dx = x - 3.5;
        const dy = y - 3.5;
        const dz = z - 3.5;
        if (Math.sqrt(dx * dx + dy * dy + dz * dz) < 3) {
          setVoxel(voxels, x, y, z, red);
        }
      }
    }
  }

  // Darker bottom
  for (let x = 2; x < 6; x++) {
    for (let z = 2; z < 6; z++) {
      if (voxels[x][1]?.[z]) setVoxel(voxels, x, 1, z, darkRed);
    }
  }

  // Stem
  setVoxel(voxels, 3, 7, 3, stem);
  setVoxel(voxels, 3, 8, 3, stem);

  // Leaf
  setVoxel(voxels, 4, 8, 3, leaf);
  setVoxel(voxels, 5, 8, 3, leaf);
  setVoxel(voxels, 4, 9, 3, leaf);

  return createVoxelGeometry(voxels, size);
}

/**
 * Create a Diamond model
 */
export function createDiamondModel(size = 1) {
  const voxels = createEmptyVoxelArray(8, 12, 8);
  const diamond = "#00FFFF";
  const highlight = "#AFFFFF";

  // Top pyramid
  for (let y = 6; y < 12; y++) {
    const layer = y - 6;
    const start = layer;
    const end = 7 - layer;
    if (start <= end) {
      for (let x = start; x <= end; x++) {
        for (let z = start; z <= end; z++) {
          setVoxel(voxels, x, y, z, diamond);
        }
      }
    }
  }

  // Bottom pyramid
  for (let y = 0; y < 6; y++) {
    const layer = 5 - y;
    const start = layer;
    const end = 7 - layer;
    if (start <= end) {
      for (let x = start; x <= end; x++) {
        for (let z = start; z <= end; z++) {
          setVoxel(voxels, x, y, z, diamond);
        }
      }
    }
  }

  // Highlight on top faces
  for (let x = 1; x < 7; x++) {
    for (let z = 1; z < 7; z++) {
      if (voxels[x][10]?.[z]) setVoxel(voxels, x, 10, z, highlight);
    }
  }

  return createVoxelGeometry(voxels, size);
}

// Color utility
function lightenColor(hex, amount) {
  const num = parseInt(hex.slice(1), 16);
  const r = Math.min(255, ((num >> 16) & 0xff) + 255 * amount);
  const g = Math.min(255, ((num >> 8) & 0xff) + 255 * amount);
  const b = Math.min(255, (num & 0xff) + 255 * amount);
  return `#${(
    (1 << 24) +
    (Math.floor(r) << 16) +
    (Math.floor(g) << 8) +
    Math.floor(b)
  )
    .toString(16)
    .slice(1)}`;
}

export default {
  createSwordModel,
  createPickaxeModel,
  createAxeModel,
  createShovelModel,
  createAppleModel,
  createDiamondModel,
};
