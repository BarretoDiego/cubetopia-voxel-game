/**
 * Gerador de Terreno Procedural com Biomas
 */

import { SimplexNoise } from "../../core/noise/SimplexNoise.js";
import { FBM } from "../../core/noise/FBM.js";
import { BlockTypes } from "../../core/blocks/BlockTypes.js";
import {
  CHUNK_SIZE,
  CHUNK_HEIGHT,
  SEA_LEVEL,
  TERRAIN_BASE_HEIGHT,
  TERRAIN_AMPLITUDE,
} from "../../utils/constants.js";
import { clamp, SeededRNG } from "../../utils/mathUtils.js";

export class TerrainGenerator {
  constructor(seed = Date.now()) {
    this.seed = seed;
    this.rng = new SeededRNG(seed);

    // Noise para diferentes aspectos
    this.heightNoise = new SimplexNoise(seed);
    this.biomeNoise = new SimplexNoise(seed + 1000);
    this.caveNoise = new SimplexNoise(seed + 2000);
    this.detailNoise = new SimplexNoise(seed + 3000);

    // FBM configurados
    this.heightFBM = new FBM({
      octaves: 6,
      lacunarity: 2.0,
      persistence: 0.5,
      scale: 0.005,
    });

    this.biomeFBM = new FBM({
      octaves: 4,
      lacunarity: 2.0,
      persistence: 0.5,
      scale: 0.002,
    });

    this.caveFBM = new FBM({
      octaves: 3,
      lacunarity: 2.0,
      persistence: 0.5,
      scale: 0.05,
    });

    // Geradores de estruturas
    this.structureGenerators = [];
  }

  /**
   * Registra um gerador de estruturas
   */
  registerStructureGenerator(generator) {
    this.structureGenerators.push(generator);
  }

  /**
   * Gera um chunk completo
   */
  generateChunk(chunk) {
    const startX = chunk.cx * CHUNK_SIZE;
    const startZ = chunk.cz * CHUNK_SIZE;

    // Primeiro passo: terreno base
    for (let lx = 0; lx < CHUNK_SIZE; lx++) {
      for (let lz = 0; lz < CHUNK_SIZE; lz++) {
        const wx = startX + lx;
        const wz = startZ + lz;

        this._generateColumn(chunk, lx, lz, wx, wz);
      }
    }

    // Segundo passo: estruturas
    this._generateStructures(chunk, startX, startZ);

    // Terceiro passo: decorações
    this._generateDecorations(chunk, startX, startZ);

    // Quarto passo: cachoeiras
    this._generateWaterfalls(chunk, startX, startZ);

    chunk.isGenerated = true;
  }

  /**
   * Gera uma coluna vertical de blocos
   */
  _generateColumn(chunk, lx, lz, wx, wz) {
    // Determina bioma
    const biome = this._getBiome(wx, wz);

    // Calcula altura do terreno
    const baseHeight = this._getTerrainHeight(wx, wz, biome);

    // Atualiza height map
    chunk.heightMap[lx + lz * CHUNK_SIZE] = baseHeight;

    for (let y = 0; y < CHUNK_HEIGHT; y++) {
      let blockType = BlockTypes.AIR;

      // Bedrock
      if (y === 0) {
        blockType = BlockTypes.BEDROCK;
      }
      // Abaixo da superfície
      else if (y < baseHeight - 4) {
        blockType = this._getUndergroundBlock(wx, y, wz, biome);
      }
      // Camada de transição
      else if (y < baseHeight) {
        blockType = this._getSubsurfaceBlock(biome);
      }
      // Superfície
      else if (y === baseHeight) {
        blockType = this._getSurfaceBlock(y, biome);
      }
      // Água
      else if (y < SEA_LEVEL && biome.hasWater) {
        blockType = BlockTypes.WATER;
      }

      if (blockType !== BlockTypes.AIR) {
        chunk.setBlock(lx, y, lz, blockType);
      }
    }
  }

  /**
   * Obtém bioma para uma posição
   */
  _getBiome(wx, wz) {
    // Temperatura e umidade
    const temperature = this.biomeFBM.sample2D(this.biomeNoise, wx, wz);
    const humidity = this.biomeFBM.sample2D(
      this.biomeNoise,
      wx + 5000,
      wz + 5000
    );

    // Determina bioma baseado em temperatura e umidade
    if (temperature > 0.3) {
      if (humidity < -0.2) {
        return {
          name: "desert",
          surface: BlockTypes.SAND,
          subsurface: BlockTypes.SAND,
          heightMod: 0.3,
          hasWater: true,
          hasTrees: false,
          hasFlowers: false,
          hasCactus: true,
        };
      } else {
        return {
          name: "plains",
          surface: BlockTypes.GRASS,
          subsurface: BlockTypes.DIRT,
          heightMod: 0.5,
          hasWater: true,
          hasTrees: true,
          treeChance: 0.01,
          hasFlowers: true,
          treType: "oak",
        };
      }
    } else if (temperature < -0.3) {
      return {
        name: "snow",
        surface: BlockTypes.SNOW,
        subsurface: BlockTypes.DIRT,
        heightMod: 0.7,
        hasWater: true,
        waterType: BlockTypes.ICE,
        hasTrees: true,
        treeChance: 0.02,
        treeType: "spruce",
      };
    } else {
      if (humidity > 0.2) {
        return {
          name: "forest",
          surface: BlockTypes.GRASS,
          subsurface: BlockTypes.DIRT,
          heightMod: 0.6,
          hasWater: true,
          hasTrees: true,
          treeChance: 0.08,
          hasFlowers: true,
          treeType: "oak",
        };
      } else {
        return {
          name: "mountains",
          surface: BlockTypes.STONE,
          subsurface: BlockTypes.STONE,
          heightMod: 1.5,
          hasWater: true,
          hasTrees: true,
          treeChance: 0.005,
          treeType: "spruce",
        };
      }
    }
  }

  /**
   * Calcula altura do terreno
   */
  _getTerrainHeight(wx, wz, biome) {
    let height = TERRAIN_BASE_HEIGHT;

    // FBM para terreno geral
    const fbmValue = this.heightFBM.sample2D(this.heightNoise, wx, wz);
    height += fbmValue * TERRAIN_AMPLITUDE * biome.heightMod;

    // Detalhe de alta frequência
    const detail = this.detailNoise.noise2D(wx * 0.1, wz * 0.1) * 2;
    height += detail;

    // Para montanhas, adiciona ridged noise
    if (biome.name === "mountains") {
      const ridged = this.heightFBM.ridged2D(this.heightNoise, wx * 2, wz * 2);
      height += ridged * 20;
    }

    return Math.floor(clamp(height, 1, CHUNK_HEIGHT - 10));
  }

  /**
   * Bloco subterrâneo
   */
  _getUndergroundBlock(wx, y, wz, biome) {
    // Cavernas
    const caveValue = this.caveFBM.sample3D(this.caveNoise, wx, y, wz);
    if (caveValue > 0.6 && y > 5) {
      return BlockTypes.AIR;
    }

    // Minérios (quanto mais fundo, mais raros)
    const oreChance = this.detailNoise.noise3D(wx * 0.2, y * 0.2, wz * 0.2);

    if (y < 15 && oreChance > 0.85) {
      return BlockTypes.DIAMOND_ORE;
    } else if (y < 30 && oreChance > 0.8) {
      return BlockTypes.GOLD_ORE;
    } else if (y < 45 && oreChance > 0.75) {
      return BlockTypes.IRON_ORE;
    } else if (oreChance > 0.7) {
      return BlockTypes.COAL_ORE;
    }

    return BlockTypes.STONE;
  }

  /**
   * Bloco de subsuperfície
   */
  _getSubsurfaceBlock(biome) {
    return biome.subsurface;
  }

  /**
   * Bloco de superfície
   */
  _getSurfaceBlock(height, biome) {
    if (height <= SEA_LEVEL + 2 && biome.name !== "desert") {
      return BlockTypes.SAND;
    }
    return biome.surface;
  }

  /**
   * Gera estruturas (árvores, etc.)
   */
  _generateStructures(chunk, startX, startZ) {
    const chunkRng = new SeededRNG(this.seed + chunk.cx * 1000 + chunk.cz);

    for (let lx = 2; lx < CHUNK_SIZE - 2; lx++) {
      for (let lz = 2; lz < CHUNK_SIZE - 2; lz++) {
        const wx = startX + lx;
        const wz = startZ + lz;
        const height = chunk.getHeight(lx, lz);

        if (height <= SEA_LEVEL) continue;

        const biome = this._getBiome(wx, wz);

        // Árvores
        if (biome.hasTrees && chunkRng.next() < biome.treeChance) {
          this._generateTree(
            chunk,
            lx,
            height + 1,
            lz,
            biome.treeType || "oak",
            chunkRng
          );
        }

        // Cactos
        if (biome.hasCactus && chunkRng.next() < 0.005) {
          this._generateCactus(chunk, lx, height + 1, lz, chunkRng);
        }
      }
    }
  }

  /**
   * Gera uma árvore
   */
  _generateTree(chunk, lx, ly, lz, type, rng) {
    const height = 4 + rng.nextInt(0, 2);

    let logType = BlockTypes.OAK_LOG;
    let leafType = BlockTypes.OAK_LEAVES;

    if (type === "birch") {
      logType = BlockTypes.BIRCH_LOG;
      leafType = BlockTypes.BIRCH_LEAVES;
    } else if (type === "spruce") {
      logType = BlockTypes.SPRUCE_LOG;
      leafType = BlockTypes.SPRUCE_LEAVES;
    }

    // Tronco
    for (let i = 0; i < height; i++) {
      if (ly + i < CHUNK_HEIGHT) {
        chunk.setBlock(lx, ly + i, lz, logType);
      }
    }

    // Folhas (copa)
    const leafStart = height - 2;
    for (let dy = leafStart; dy <= height + 1; dy++) {
      const radius = dy === height + 1 ? 1 : 2;
      for (let dx = -radius; dx <= radius; dx++) {
        for (let dz = -radius; dz <= radius; dz++) {
          if (Math.abs(dx) + Math.abs(dz) <= radius + 1) {
            const nlx = lx + dx;
            const nlz = lz + dz;
            const nly = ly + dy;

            if (
              nlx >= 0 &&
              nlx < CHUNK_SIZE &&
              nlz >= 0 &&
              nlz < CHUNK_SIZE &&
              nly < CHUNK_HEIGHT
            ) {
              if (chunk.getBlock(nlx, nly, nlz) === BlockTypes.AIR) {
                chunk.setBlock(nlx, nly, nlz, leafType);
              }
            }
          }
        }
      }
    }
  }

  /**
   * Gera um cacto
   */
  _generateCactus(chunk, lx, ly, lz, rng) {
    const height = 2 + rng.nextInt(0, 2);

    for (let i = 0; i < height; i++) {
      if (ly + i < CHUNK_HEIGHT) {
        chunk.setBlock(lx, ly + i, lz, BlockTypes.CACTUS);
      }
    }
  }

  /**
   * Gera decorações (flores, grama alta, etc.)
   */
  _generateDecorations(chunk, startX, startZ) {
    const chunkRng = new SeededRNG(this.seed + chunk.cx * 2000 + chunk.cz);

    for (let lx = 0; lx < CHUNK_SIZE; lx++) {
      for (let lz = 0; lz < CHUNK_SIZE; lz++) {
        const wx = startX + lx;
        const wz = startZ + lz;
        const height = chunk.getHeight(lx, lz);

        if (height <= SEA_LEVEL) continue;

        const biome = this._getBiome(wx, wz);
        const surfaceBlock = chunk.getBlock(lx, height, lz);

        if (surfaceBlock !== BlockTypes.GRASS) continue;

        if (biome.hasFlowers) {
          // Grama alta
          if (chunkRng.next() < 0.15) {
            chunk.setBlock(lx, height + 1, lz, BlockTypes.TALL_GRASS);
          }
          // Flores
          else if (chunkRng.next() < 0.02) {
            const flowerType =
              chunkRng.next() > 0.5
                ? BlockTypes.FLOWER_RED
                : BlockTypes.FLOWER_YELLOW;
            chunk.setBlock(lx, height + 1, lz, flowerType);
          }
        }

        // Cogumelos (raro, em áreas escuras)
        if (chunkRng.next() < 0.005) {
          const mushroom =
            chunkRng.next() > 0.5
              ? BlockTypes.MUSHROOM_RED
              : BlockTypes.MUSHROOM_BROWN;
          chunk.setBlock(lx, height + 1, lz, mushroom);
        }
      }
    }
  }

  /**
   * Gera cachoeiras em biomas montanhosos
   * Procura por penhascos e coloca água no topo
   */
  _generateWaterfalls(chunk, startX, startZ) {
    const chunkRng = new SeededRNG(this.seed + chunk.cx * 3000 + chunk.cz);

    // Só gera cachoeiras ocasionalmente
    if (chunkRng.next() > 0.15) return; // 15% chance por chunk

    for (let lx = 3; lx < CHUNK_SIZE - 3; lx++) {
      for (let lz = 3; lz < CHUNK_SIZE - 3; lz++) {
        const wx = startX + lx;
        const wz = startZ + lz;
        const biome = this._getBiome(wx, wz);

        // Só em montanhas
        if (biome.name !== "mountains") continue;

        const height = chunk.getHeight(lx, lz);

        // Precisa estar em altura elevada
        if (height < 35) continue;

        // Verifica se há um penhasco (grande diferença de altura)
        const directions = [
          { dx: 1, dz: 0 },
          { dx: -1, dz: 0 },
          { dx: 0, dz: 1 },
          { dx: 0, dz: -1 },
        ];

        for (const { dx, dz } of directions) {
          const neighborHeight = chunk.getHeight(lx + dx * 2, lz + dz * 2);
          const heightDiff = height - neighborHeight;

          // Precisa de uma queda significativa (pelo menos 8 blocos)
          if (heightDiff >= 8 && chunkRng.next() < 0.3) {
            // Coloca fonte de água no topo do penhasco
            chunk.setBlock(lx, height, lz, BlockTypes.WATER);

            // Cria a cascata descendo o penhasco
            let currentY = height - 1;
            let currentX = lx + dx;
            let currentZ = lz + dz;

            while (currentY > neighborHeight && currentY > SEA_LEVEL) {
              if (
                currentX >= 0 &&
                currentX < CHUNK_SIZE &&
                currentZ >= 0 &&
                currentZ < CHUNK_SIZE
              ) {
                const blockBelow = chunk.getBlock(currentX, currentY, currentZ);

                // Se há ar, coloca água
                if (blockBelow === BlockTypes.AIR) {
                  chunk.setBlock(
                    currentX,
                    currentY,
                    currentZ,
                    BlockTypes.WATER
                  );
                }
              }
              currentY--;
            }

            // Cria uma pequena poça na base
            if (
              currentX >= 1 &&
              currentX < CHUNK_SIZE - 1 &&
              currentZ >= 1 &&
              currentZ < CHUNK_SIZE - 1 &&
              currentY >= SEA_LEVEL
            ) {
              for (let px = -1; px <= 1; px++) {
                for (let pz = -1; pz <= 1; pz++) {
                  const poolX = currentX + px;
                  const poolZ = currentZ + pz;
                  if (
                    poolX >= 0 &&
                    poolX < CHUNK_SIZE &&
                    poolZ >= 0 &&
                    poolZ < CHUNK_SIZE
                  ) {
                    const groundHeight = chunk.getHeight(poolX, poolZ);
                    if (
                      chunk.getBlock(poolX, groundHeight + 1, poolZ) ===
                      BlockTypes.AIR
                    ) {
                      chunk.setBlock(
                        poolX,
                        groundHeight + 1,
                        poolZ,
                        BlockTypes.WATER
                      );
                    }
                  }
                }
              }
            }

            // Só uma cachoeira por chunk
            return;
          }
        }
      }
    }
  }
}

export default TerrainGenerator;
