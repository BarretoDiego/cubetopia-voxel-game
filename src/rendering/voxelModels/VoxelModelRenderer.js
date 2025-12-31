/**
 * High-Resolution Voxel Model Renderer
 * Creates detailed 3D voxel models using smaller voxel units
 */

import * as THREE from "three";

// Voxel size for high-res models (smaller = more detail)
const VOXEL_SIZE = 0.0625; // 1/16 of a block

/**
 * Creates a geometry from a 3D voxel array
 * @param {Array} voxels - 3D array of colors (null = empty)
 * @param {number} scale - Scale multiplier
 */
export function createVoxelGeometry(voxels, scale = 1) {
  const positions = [];
  const colors = [];
  const normals = [];
  const indices = [];

  const sizeX = voxels.length;
  const sizeY = voxels[0]?.length || 0;
  const sizeZ = voxels[0]?.[0]?.length || 0;

  // Center offset
  const offsetX = (sizeX * VOXEL_SIZE * scale) / 2;
  const offsetY = 0;
  const offsetZ = (sizeZ * VOXEL_SIZE * scale) / 2;

  let vertexIndex = 0;

  // Face definitions
  const faces = [
    {
      dir: [0, 1, 0],
      corners: [
        [0, 1, 1],
        [1, 1, 1],
        [1, 1, 0],
        [0, 1, 0],
      ],
    }, // top
    {
      dir: [0, -1, 0],
      corners: [
        [0, 0, 0],
        [1, 0, 0],
        [1, 0, 1],
        [0, 0, 1],
      ],
    }, // bottom
    {
      dir: [0, 0, 1],
      corners: [
        [0, 0, 1],
        [1, 0, 1],
        [1, 1, 1],
        [0, 1, 1],
      ],
    }, // front
    {
      dir: [0, 0, -1],
      corners: [
        [1, 0, 0],
        [0, 0, 0],
        [0, 1, 0],
        [1, 1, 0],
      ],
    }, // back
    {
      dir: [-1, 0, 0],
      corners: [
        [0, 0, 0],
        [0, 0, 1],
        [0, 1, 1],
        [0, 1, 0],
      ],
    }, // left
    {
      dir: [1, 0, 0],
      corners: [
        [1, 0, 1],
        [1, 0, 0],
        [1, 1, 0],
        [1, 1, 1],
      ],
    }, // right
  ];

  for (let x = 0; x < sizeX; x++) {
    for (let y = 0; y < sizeY; y++) {
      for (let z = 0; z < sizeZ; z++) {
        const voxel = voxels[x][y][z];
        if (!voxel) continue;

        const color = hexToRgb(voxel);

        // Check each face
        for (const face of faces) {
          const [dx, dy, dz] = face.dir;
          const nx = x + dx;
          const ny = y + dy;
          const nz = z + dz;

          // Only add face if neighbor is empty
          const neighborEmpty =
            nx < 0 ||
            nx >= sizeX ||
            ny < 0 ||
            ny >= sizeY ||
            nz < 0 ||
            nz >= sizeZ ||
            !voxels[nx][ny][nz];

          if (neighborEmpty) {
            // Add face vertices
            for (const corner of face.corners) {
              const px = (x + corner[0]) * VOXEL_SIZE * scale - offsetX;
              const py = (y + corner[1]) * VOXEL_SIZE * scale - offsetY;
              const pz = (z + corner[2]) * VOXEL_SIZE * scale - offsetZ;

              positions.push(px, py, pz);
              colors.push(color.r, color.g, color.b);
              normals.push(dx, dy, dz);
            }

            // Add indices for two triangles
            indices.push(
              vertexIndex,
              vertexIndex + 1,
              vertexIndex + 2,
              vertexIndex,
              vertexIndex + 2,
              vertexIndex + 3
            );
            vertexIndex += 4;
          }
        }
      }
    }
  }

  if (positions.length === 0) return null;

  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute(
    "position",
    new THREE.Float32BufferAttribute(positions, 3)
  );
  geometry.setAttribute("color", new THREE.Float32BufferAttribute(colors, 3));
  geometry.setAttribute("normal", new THREE.Float32BufferAttribute(normals, 3));
  geometry.setIndex(indices);

  return geometry;
}

/**
 * Convert hex color to RGB (0-1 range)
 */
function hexToRgb(hex) {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result
    ? {
        r: parseInt(result[1], 16) / 255,
        g: parseInt(result[2], 16) / 255,
        b: parseInt(result[3], 16) / 255,
      }
    : { r: 1, g: 0, b: 1 }; // Magenta for invalid
}

/**
 * Create a 3D voxel array filled with a color
 */
export function createFilledVoxelArray(sizeX, sizeY, sizeZ, color) {
  const arr = [];
  for (let x = 0; x < sizeX; x++) {
    arr[x] = [];
    for (let y = 0; y < sizeY; y++) {
      arr[x][y] = [];
      for (let z = 0; z < sizeZ; z++) {
        arr[x][y][z] = color;
      }
    }
  }
  return arr;
}

/**
 * Create an empty 3D voxel array
 */
export function createEmptyVoxelArray(sizeX, sizeY, sizeZ) {
  const arr = [];
  for (let x = 0; x < sizeX; x++) {
    arr[x] = [];
    for (let y = 0; y < sizeY; y++) {
      arr[x][y] = new Array(sizeZ).fill(null);
    }
  }
  return arr;
}

/**
 * Set a voxel in a 3D array
 */
export function setVoxel(arr, x, y, z, color) {
  if (arr[x] && arr[x][y]) {
    arr[x][y][z] = color;
  }
}

/**
 * Fill a box region in a voxel array
 */
export function fillBox(arr, x1, y1, z1, x2, y2, z2, color) {
  for (let x = x1; x <= x2; x++) {
    for (let y = y1; y <= y2; y++) {
      for (let z = z1; z <= z2; z++) {
        setVoxel(arr, x, y, z, color);
      }
    }
  }
}

export default {
  createVoxelGeometry,
  createFilledVoxelArray,
  createEmptyVoxelArray,
  setVoxel,
  fillBox,
  VOXEL_SIZE,
};
