/**
 * Entity Models - High-resolution voxel models for creatures
 */

import {
  createEmptyVoxelArray,
  fillBox,
  setVoxel,
  createVoxelGeometry,
} from "./VoxelModelRenderer.js";

/**
 * Create a Slime model - jiggly cube with face
 */
export function createSlimeModel(color = "#22CC22", size = 1) {
  const resolution = 16;
  const voxels = createEmptyVoxelArray(resolution, resolution, resolution);

  // Lighter and darker versions of color
  const darkColor = darkenColor(color, 0.3);
  const lightColor = lightenColor(color, 0.2);

  // Main body - rounded cube
  for (let x = 2; x < 14; x++) {
    for (let y = 1; y < 13; y++) {
      for (let z = 2; z < 14; z++) {
        // Create rounded corners
        const dx = Math.abs(x - 8);
        const dy = Math.abs(y - 7);
        const dz = Math.abs(z - 8);

        if (dx + dy + dz < 14) {
          setVoxel(voxels, x, y, z, color);
        }
      }
    }
  }

  // Eyes - white with black pupils
  // Left eye
  fillBox(voxels, 5, 8, 13, 6, 10, 14, "#FFFFFF");
  setVoxel(voxels, 5, 9, 14, "#000000");
  setVoxel(voxels, 6, 9, 14, "#000000");

  // Right eye
  fillBox(voxels, 9, 8, 13, 10, 10, 14, "#FFFFFF");
  setVoxel(voxels, 9, 9, 14, "#000000");
  setVoxel(voxels, 10, 9, 14, "#000000");

  // Mouth
  fillBox(voxels, 6, 6, 14, 9, 6, 14, darkColor);

  // Shading - darker bottom
  for (let x = 2; x < 14; x++) {
    for (let z = 2; z < 14; z++) {
      if (voxels[x][1]?.[z]) setVoxel(voxels, x, 1, z, darkColor);
      if (voxels[x][2]?.[z]) setVoxel(voxels, x, 2, z, darkColor);
    }
  }

  // Highlight - lighter top
  for (let x = 4; x < 12; x++) {
    for (let z = 4; z < 12; z++) {
      if (voxels[x][11]?.[z]) setVoxel(voxels, x, 11, z, lightColor);
      if (voxels[x][12]?.[z]) setVoxel(voxels, x, 12, z, lightColor);
    }
  }

  return createVoxelGeometry(voxels, size);
}

/**
 * Create a Pig model
 */
export function createPigModel(size = 1) {
  const voxels = createEmptyVoxelArray(16, 16, 24);
  const pink = "#FFB6C1";
  const darkPink = "#FF69B4";
  const black = "#000000";

  // Body
  fillBox(voxels, 3, 4, 6, 12, 11, 18, pink);

  // Head
  fillBox(voxels, 4, 6, 18, 11, 13, 24, pink);

  // Snout
  fillBox(voxels, 6, 7, 23, 9, 10, 24, darkPink);
  setVoxel(voxels, 6, 8, 24, black);
  setVoxel(voxels, 9, 8, 24, black);

  // Eyes
  setVoxel(voxels, 5, 11, 23, black);
  setVoxel(voxels, 10, 11, 23, black);

  // Ears
  fillBox(voxels, 4, 12, 20, 5, 14, 22, pink);
  fillBox(voxels, 10, 12, 20, 11, 14, 22, pink);

  // Legs
  fillBox(voxels, 4, 0, 7, 5, 4, 9, pink);
  fillBox(voxels, 10, 0, 7, 11, 4, 9, pink);
  fillBox(voxels, 4, 0, 15, 5, 4, 17, pink);
  fillBox(voxels, 10, 0, 15, 11, 4, 17, pink);

  // Tail (curly)
  setVoxel(voxels, 7, 9, 5, darkPink);
  setVoxel(voxels, 7, 10, 4, darkPink);
  setVoxel(voxels, 8, 10, 4, darkPink);

  return createVoxelGeometry(voxels, size);
}

/**
 * Create a Zombie model
 */
export function createZombieModel(size = 1) {
  const voxels = createEmptyVoxelArray(16, 32, 8);
  const skinGreen = "#5A8A5A";
  const skinDark = "#486B48";
  const shirt = "#4169E1";
  const pants = "#483D8B";
  const black = "#000000";

  // Legs
  fillBox(voxels, 4, 0, 2, 6, 11, 5, pants);
  fillBox(voxels, 9, 0, 2, 11, 11, 5, pants);

  // Torso
  fillBox(voxels, 4, 12, 2, 11, 23, 5, shirt);
  // Torn shirt effect
  setVoxel(voxels, 5, 13, 5, null);
  setVoxel(voxels, 10, 15, 5, null);
  setVoxel(voxels, 6, 14, 5, skinGreen);

  // Arms
  fillBox(voxels, 1, 12, 2, 3, 23, 5, skinGreen);
  fillBox(voxels, 12, 12, 2, 14, 23, 5, skinGreen);
  // Arms extended forward
  fillBox(voxels, 1, 20, 5, 3, 23, 7, skinGreen);
  fillBox(voxels, 12, 20, 5, 14, 23, 7, skinGreen);

  // Head
  fillBox(voxels, 4, 24, 1, 11, 31, 6, skinGreen);

  // Face details
  // Eyes
  setVoxel(voxels, 5, 28, 6, black);
  setVoxel(voxels, 10, 28, 6, black);
  // Mouth
  fillBox(voxels, 6, 25, 6, 9, 25, 6, skinDark);

  // Hair
  fillBox(voxels, 4, 30, 2, 11, 31, 5, skinDark);

  return createVoxelGeometry(voxels, size);
}

/**
 * Create a Spider model
 */
export function createSpiderModel(size = 1) {
  const voxels = createEmptyVoxelArray(24, 8, 16);
  const body = "#2F1F1F";
  const eyes = "#FF0000";

  // Abdomen
  fillBox(voxels, 9, 2, 1, 14, 6, 7, body);

  // Thorax
  fillBox(voxels, 10, 2, 8, 13, 5, 12, body);

  // Head
  fillBox(voxels, 10, 2, 12, 13, 5, 15, body);

  // Eyes (8 eyes in 2 rows)
  setVoxel(voxels, 10, 4, 15, eyes);
  setVoxel(voxels, 11, 4, 15, eyes);
  setVoxel(voxels, 12, 4, 15, eyes);
  setVoxel(voxels, 13, 4, 15, eyes);
  setVoxel(voxels, 10, 3, 15, eyes);
  setVoxel(voxels, 11, 3, 15, eyes);
  setVoxel(voxels, 12, 3, 15, eyes);
  setVoxel(voxels, 13, 3, 15, eyes);

  // Legs (8 legs)
  // Left side
  fillBox(voxels, 5, 1, 9, 9, 2, 9, body);
  fillBox(voxels, 4, 1, 10, 9, 2, 10, body);
  fillBox(voxels, 3, 1, 11, 9, 2, 11, body);
  fillBox(voxels, 2, 1, 12, 9, 2, 12, body);

  // Right side
  fillBox(voxels, 14, 1, 9, 18, 2, 9, body);
  fillBox(voxels, 14, 1, 10, 19, 2, 10, body);
  fillBox(voxels, 14, 1, 11, 20, 2, 11, body);
  fillBox(voxels, 14, 1, 12, 21, 2, 12, body);

  // Leg ends touch ground
  for (let i = 0; i < 4; i++) {
    setVoxel(voxels, 2 + i, 0, 9 + i, body);
    setVoxel(voxels, 21 - i, 0, 9 + i, body);
  }

  return createVoxelGeometry(voxels, size);
}

/**
 * Create a Bird model
 */
export function createBirdModel(color = "#FF6347", size = 1) {
  const voxels = createEmptyVoxelArray(16, 8, 12);
  const black = "#000000";
  const beak = "#FFA500";

  // Body
  fillBox(voxels, 5, 2, 3, 10, 5, 9, color);

  // Head
  fillBox(voxels, 6, 4, 9, 9, 7, 11, color);

  // Eyes
  setVoxel(voxels, 6, 6, 11, black);
  setVoxel(voxels, 9, 6, 11, black);

  // Beak
  fillBox(voxels, 7, 5, 11, 8, 5, 12, beak);

  // Wings
  fillBox(voxels, 2, 3, 4, 4, 4, 8, color);
  fillBox(voxels, 11, 3, 4, 13, 4, 8, color);

  // Tail
  fillBox(voxels, 6, 3, 1, 9, 4, 3, color);

  // Feet
  setVoxel(voxels, 6, 1, 5, beak);
  setVoxel(voxels, 9, 1, 5, beak);

  return createVoxelGeometry(voxels, size);
}

/**
 * Create a Cow model
 */
export function createCowModel(size = 1) {
  const voxels = createEmptyVoxelArray(16, 20, 28);
  const white = "#FFFFFF";
  const black = "#1A1A1A";
  const pink = "#FFB6C1";
  const horn = "#F5DEB3";

  // Body
  fillBox(voxels, 3, 6, 6, 12, 14, 22, white);
  // Black spots
  fillBox(voxels, 4, 8, 8, 6, 12, 12, black);
  fillBox(voxels, 9, 10, 14, 11, 13, 18, black);
  fillBox(voxels, 5, 7, 18, 7, 10, 20, black);

  // Head
  fillBox(voxels, 4, 10, 22, 11, 18, 27, white);

  // Snout
  fillBox(voxels, 5, 10, 26, 10, 14, 27, pink);
  setVoxel(voxels, 6, 12, 27, black);
  setVoxel(voxels, 9, 12, 27, black);

  // Eyes
  setVoxel(voxels, 5, 16, 26, black);
  setVoxel(voxels, 10, 16, 26, black);

  // Horns
  fillBox(voxels, 3, 17, 23, 4, 19, 24, horn);
  fillBox(voxels, 11, 17, 23, 12, 19, 24, horn);

  // Ears
  fillBox(voxels, 3, 15, 22, 4, 17, 23, white);
  fillBox(voxels, 11, 15, 22, 12, 17, 23, white);

  // Legs
  fillBox(voxels, 4, 0, 7, 5, 6, 9, white);
  fillBox(voxels, 10, 0, 7, 11, 6, 9, white);
  fillBox(voxels, 4, 0, 19, 5, 6, 21, white);
  fillBox(voxels, 10, 0, 19, 11, 6, 21, white);

  // Udder
  fillBox(voxels, 6, 5, 14, 9, 6, 17, pink);

  // Tail
  fillBox(voxels, 7, 10, 4, 8, 12, 6, white);
  fillBox(voxels, 7, 8, 3, 8, 10, 5, black);

  return createVoxelGeometry(voxels, size);
}

// Color utility functions
function darkenColor(hex, amount) {
  const num = parseInt(hex.slice(1), 16);
  const r = Math.max(0, ((num >> 16) & 0xff) * (1 - amount));
  const g = Math.max(0, ((num >> 8) & 0xff) * (1 - amount));
  const b = Math.max(0, (num & 0xff) * (1 - amount));
  return `#${(
    (1 << 24) +
    (Math.floor(r) << 16) +
    (Math.floor(g) << 8) +
    Math.floor(b)
  )
    .toString(16)
    .slice(1)}`;
}

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
  createSlimeModel,
  createPigModel,
  createZombieModel,
  createSpiderModel,
  createBirdModel,
  createCowModel,
};
