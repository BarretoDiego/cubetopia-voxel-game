/**
 * ChunkMesher - Gera meshes otimizadas para chunks
 */

import * as THREE from "three";
import { CHUNK_SIZE, CHUNK_HEIGHT } from "../../utils/constants.js";
import { BlockTypes } from "../blocks/BlockTypes.js";
import { blockRegistry } from "../blocks/BlockRegistry.js";

// Vértices para cada face de um cubo
const FACE_VERTICES = {
  top: [
    [0, 1, 0],
    [1, 1, 0],
    [1, 1, 1],
    [0, 1, 1],
  ],
  bottom: [
    [0, 0, 1],
    [1, 0, 1],
    [1, 0, 0],
    [0, 0, 0],
  ],
  front: [
    [0, 0, 1],
    [0, 1, 1],
    [1, 1, 1],
    [1, 0, 1],
  ],
  back: [
    [1, 0, 0],
    [1, 1, 0],
    [0, 1, 0],
    [0, 0, 0],
  ],
  left: [
    [0, 0, 0],
    [0, 1, 0],
    [0, 1, 1],
    [0, 0, 1],
  ],
  right: [
    [1, 0, 1],
    [1, 1, 1],
    [1, 1, 0],
    [1, 0, 0],
  ],
};

// Normais para cada face
const FACE_NORMALS = {
  top: [0, 1, 0],
  bottom: [0, -1, 0],
  front: [0, 0, 1],
  back: [0, 0, -1],
  left: [-1, 0, 0],
  right: [1, 0, 0],
};

// UVs para as faces
const FACE_UVS = [
  [0, 0],
  [1, 0],
  [1, 1],
  [0, 1],
];

export class ChunkMesher {
  constructor(textureManager = null) {
    this.textureManager = textureManager;

    // Buffers reutilizáveis
    this.positions = [];
    this.normals = [];
    this.uvs = [];
    this.colors = [];
    this.indices = [];
    this.aoValues = [];
  }

  /**
   * Gera mesh para um chunk
   */
  generateMesh(chunk, getBlockFunc) {
    this._resetBuffers();

    const worldOffsetX = chunk.cx * CHUNK_SIZE;
    const worldOffsetZ = chunk.cz * CHUNK_SIZE;

    // Itera sobre todos os blocos
    for (let z = 0; z < CHUNK_SIZE; z++) {
      for (let y = 0; y < CHUNK_HEIGHT; y++) {
        for (let x = 0; x < CHUNK_SIZE; x++) {
          const blockType = chunk.getBlock(x, y, z);

          if (blockType === BlockTypes.AIR) continue;

          const blockDef = blockRegistry.get(blockType);
          const worldX = worldOffsetX + x;
          const worldZ = worldOffsetZ + z;

          // Verifica cada face
          this._addVisibleFaces(
            x,
            y,
            z,
            worldX,
            y,
            worldZ,
            blockType,
            blockDef,
            chunk,
            getBlockFunc
          );
        }
      }
    }

    return this._createGeometry();
  }

  /**
   * Adiciona faces visíveis de um bloco
   */
  _addVisibleFaces(
    lx,
    ly,
    lz,
    wx,
    wy,
    wz,
    blockType,
    blockDef,
    chunk,
    getBlockFunc
  ) {
    const faces = ["top", "bottom", "front", "back", "left", "right"];
    const neighbors = [
      [0, 1, 0], // top
      [0, -1, 0], // bottom
      [0, 0, 1], // front
      [0, 0, -1], // back
      [-1, 0, 0], // left
      [1, 0, 0], // right
    ];

    for (let i = 0; i < 6; i++) {
      const [dx, dy, dz] = neighbors[i];
      const neighborType = getBlockFunc(wx + dx, wy + dy, wz + dz);

      // Só adiciona face se vizinho for ar ou transparente
      const neighborDef = blockRegistry.get(neighborType);

      if (
        neighborType === BlockTypes.AIR ||
        (neighborDef.transparent && neighborType !== blockType)
      ) {
        this._addFace(
          faces[i],
          wx,
          wy,
          wz,
          blockDef,
          chunk,
          lx,
          ly,
          lz,
          getBlockFunc
        );
      }
    }
  }

  /**
   * Adiciona uma face ao buffer
   */
  _addFace(face, x, y, z, blockDef, chunk, lx, ly, lz, getBlockFunc) {
    const vertices = FACE_VERTICES[face];
    const normal = FACE_NORMALS[face];
    const baseIndex = this.positions.length / 3;

    // Obtém cor do bloco
    const color = this._hexToRgb(blockDef.color || "#ff00ff");

    // Adiciona 4 vértices da face
    for (let i = 0; i < 4; i++) {
      const [vx, vy, vz] = vertices[i];

      // Posição
      this.positions.push(x + vx, y + vy, z + vz);

      // Normal
      this.normals.push(...normal);

      // UV
      this.uvs.push(FACE_UVS[i][0], FACE_UVS[i][1]);

      // Cor com AO
      const ao = this._calculateAO(
        lx + vx,
        ly + vy,
        lz + vz,
        face,
        chunk,
        getBlockFunc
      );
      const aoFactor = 1.0 - ao * 0.2;

      this.colors.push(
        color.r * aoFactor,
        color.g * aoFactor,
        color.b * aoFactor
      );
    }

    // Índices (dois triângulos)
    this.indices.push(
      baseIndex,
      baseIndex + 1,
      baseIndex + 2,
      baseIndex,
      baseIndex + 2,
      baseIndex + 3
    );
  }

  /**
   * Calcula Ambient Occlusion para um vértice
   */
  _calculateAO(vx, vy, vz, face, chunk, getBlockFunc) {
    // Simplificado - conta blocos vizinhos ao vértice
    let count = 0;
    const offsets = [
      [-1, 0, 0],
      [1, 0, 0],
      [0, -1, 0],
      [0, 1, 0],
      [0, 0, -1],
      [0, 0, 1],
      [-1, -1, 0],
      [1, 1, 0],
      [-1, 0, -1],
      [1, 0, 1],
    ];

    const wx = chunk.cx * CHUNK_SIZE + Math.floor(vx);
    const wz = chunk.cz * CHUNK_SIZE + Math.floor(vz);

    for (const [dx, dy, dz] of offsets) {
      const neighbor = getBlockFunc(wx + dx, Math.floor(vy) + dy, wz + dz);
      if (
        neighbor !== BlockTypes.AIR &&
        !blockRegistry.isTransparent(neighbor)
      ) {
        count++;
      }
    }

    return Math.min(count / 4, 1);
  }

  /**
   * Converte cor hex para RGB normalizado
   */
  _hexToRgb(hex) {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
    return result
      ? {
          r: parseInt(result[1], 16) / 255,
          g: parseInt(result[2], 16) / 255,
          b: parseInt(result[3], 16) / 255,
        }
      : { r: 1, g: 0, b: 1 };
  }

  /**
   * Cria BufferGeometry a partir dos buffers
   */
  _createGeometry() {
    if (this.positions.length === 0) {
      return null;
    }

    const geometry = new THREE.BufferGeometry();

    geometry.setAttribute(
      "position",
      new THREE.Float32BufferAttribute(this.positions, 3)
    );
    geometry.setAttribute(
      "normal",
      new THREE.Float32BufferAttribute(this.normals, 3)
    );
    geometry.setAttribute("uv", new THREE.Float32BufferAttribute(this.uvs, 2));
    geometry.setAttribute(
      "color",
      new THREE.Float32BufferAttribute(this.colors, 3)
    );
    geometry.setIndex(this.indices);

    geometry.computeBoundingSphere();

    // Aggressive cleanup - clear buffers immediately after creating geometry
    this._resetBuffers();

    return geometry;
  }

  /**
   * Reseta buffers para reutilização
   */
  _resetBuffers() {
    this.positions.length = 0;
    this.normals.length = 0;
    this.uvs.length = 0;
    this.colors.length = 0;
    this.indices.length = 0;
    this.aoValues.length = 0;
  }
}

// Export singleton to avoid reallocation overhead
export const sharedChunkMesher = new ChunkMesher();
export default sharedChunkMesher;
