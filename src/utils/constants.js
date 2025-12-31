/**
 * Constantes globais do motor voxel
 */

// Dimensões dos chunks
export const CHUNK_SIZE = 16;
export const CHUNK_HEIGHT = 64;

// Distância de renderização (em chunks)
export const RENDER_DISTANCE = 3;

// Física
export const GRAVITY = 32; // Mais peso
export const JUMP_FORCE = 12; // Pulo mais rápido
export const PLAYER_SPEED = 10; // Movimento mais rápido
export const SPRINT_MULTIPLIER = 1.8;

// Geração de terreno
export const SEA_LEVEL = 12;
export const TERRAIN_BASE_HEIGHT = 20;
export const TERRAIN_AMPLITUDE = 30;

// Sementes para geração procedural
export const DEFAULT_SEED = Date.now();

// Cores de biomas
export const BIOME_COLORS = {
  plains: "#567d46",
  desert: "#e0c090",
  forest: "#2d5a27",
  mountains: "#8a8a8a",
  snow: "#f0f0f0",
  ocean: "#1a5f7a",
};

// Direções para verificação de vizinhos
export const DIRECTIONS = [
  { x: 0, y: 1, z: 0, name: "top" },
  { x: 0, y: -1, z: 0, name: "bottom" },
  { x: -1, y: 0, z: 0, name: "left" },
  { x: 1, y: 0, z: 0, name: "right" },
  { x: 0, y: 0, z: 1, name: "front" },
  { x: 0, y: 0, z: -1, name: "back" },
];

// Eixos
export const AXIS = {
  X: "x",
  Y: "y",
  Z: "z",
};
