/**
 * Classe Chunk - Representa um pedaço do mundo
 */

import { CHUNK_SIZE, CHUNK_HEIGHT } from "../../utils/constants.js";
import { BlockTypes } from "../blocks/BlockTypes.js";
import { mod } from "../../utils/mathUtils.js";

export class Chunk {
  constructor(cx, cz) {
    this.cx = cx;
    this.cz = cz;
    this.id = `${cx},${cz}`;

    // Array de blocos
    this.data = new Uint8Array(CHUNK_SIZE * CHUNK_HEIGHT * CHUNK_SIZE);

    // Flags
    this.isGenerated = false;
    this.isDirty = true;
    this.mesh = null;

    // Cache de altura para otimização
    this.heightMap = new Uint8Array(CHUNK_SIZE * CHUNK_SIZE);

    // Estatísticas
    this.solidBlockCount = 0;
  }

  /**
   * Obtém bloco em coordenadas locais
   */
  getBlock(lx, ly, lz) {
    if (
      lx < 0 ||
      lx >= CHUNK_SIZE ||
      lz < 0 ||
      lz >= CHUNK_SIZE ||
      ly < 0 ||
      ly >= CHUNK_HEIGHT
    ) {
      return BlockTypes.AIR;
    }

    const idx = this._getIndex(lx, ly, lz);
    return this.data[idx];
  }

  /**
   * Define bloco em coordenadas locais
   */
  setBlock(lx, ly, lz, type) {
    if (
      lx < 0 ||
      lx >= CHUNK_SIZE ||
      lz < 0 ||
      lz >= CHUNK_SIZE ||
      ly < 0 ||
      ly >= CHUNK_HEIGHT
    ) {
      return false;
    }

    const idx = this._getIndex(lx, ly, lz);
    const oldType = this.data[idx];

    if (oldType === type) return false;

    this.data[idx] = type;
    this.isDirty = true;

    // Atualiza contagem de blocos sólidos
    if (oldType === BlockTypes.AIR && type !== BlockTypes.AIR) {
      this.solidBlockCount++;
    } else if (oldType !== BlockTypes.AIR && type === BlockTypes.AIR) {
      this.solidBlockCount--;
    }

    // Atualiza height map
    if (type !== BlockTypes.AIR) {
      const hmIdx = lx + lz * CHUNK_SIZE;
      if (ly > this.heightMap[hmIdx]) {
        this.heightMap[hmIdx] = ly;
      }
    }

    return true;
  }

  /**
   * Obtém a altura máxima em uma coluna
   */
  getHeight(lx, lz) {
    if (lx < 0 || lx >= CHUNK_SIZE || lz < 0 || lz >= CHUNK_SIZE) {
      return 0;
    }
    return this.heightMap[lx + lz * CHUNK_SIZE];
  }

  /**
   * Converte coordenadas locais para índice no array
   */
  _getIndex(lx, ly, lz) {
    return lx + ly * CHUNK_SIZE + lz * CHUNK_SIZE * CHUNK_HEIGHT;
  }

  /**
   * Itera sobre todos os blocos não-ar
   */
  forEachSolidBlock(callback) {
    for (let z = 0; z < CHUNK_SIZE; z++) {
      for (let y = 0; y < CHUNK_HEIGHT; y++) {
        for (let x = 0; x < CHUNK_SIZE; x++) {
          const type = this.getBlock(x, y, z);
          if (type !== BlockTypes.AIR) {
            callback(x, y, z, type);
          }
        }
      }
    }
  }

  /**
   * Verifica se um bloco tem vizinhos expostos (para culling)
   */
  isExposed(lx, ly, lz) {
    return (
      this.getBlock(lx - 1, ly, lz) === BlockTypes.AIR ||
      this.getBlock(lx + 1, ly, lz) === BlockTypes.AIR ||
      this.getBlock(lx, ly - 1, lz) === BlockTypes.AIR ||
      this.getBlock(lx, ly + 1, lz) === BlockTypes.AIR ||
      this.getBlock(lx, ly, lz - 1) === BlockTypes.AIR ||
      this.getBlock(lx, ly, lz + 1) === BlockTypes.AIR
    );
  }

  /**
   * Obtém faces visíveis de um bloco
   */
  getVisibleFaces(lx, ly, lz) {
    return {
      top:
        ly === CHUNK_HEIGHT - 1 ||
        this.getBlock(lx, ly + 1, lz) === BlockTypes.AIR,
      bottom: ly === 0 || this.getBlock(lx, ly - 1, lz) === BlockTypes.AIR,
      left: lx === 0 || this.getBlock(lx - 1, ly, lz) === BlockTypes.AIR,
      right:
        lx === CHUNK_SIZE - 1 ||
        this.getBlock(lx + 1, ly, lz) === BlockTypes.AIR,
      front:
        lz === CHUNK_SIZE - 1 ||
        this.getBlock(lx, ly, lz + 1) === BlockTypes.AIR,
      back: lz === 0 || this.getBlock(lx, ly, lz - 1) === BlockTypes.AIR,
    };
  }

  /**
   * Serializa o chunk para armazenamento
   */
  serialize() {
    return {
      cx: this.cx,
      cz: this.cz,
      data: Array.from(this.data),
      heightMap: Array.from(this.heightMap),
    };
  }

  /**
   * Deserializa o chunk
   */
  static deserialize(obj) {
    const chunk = new Chunk(obj.cx, obj.cz);
    chunk.data = new Uint8Array(obj.data);
    chunk.heightMap = new Uint8Array(obj.heightMap);
    chunk.isGenerated = true;
    chunk.isDirty = true;
    return chunk;
  }

  /**
   * Limpa recursos do chunk
   */
  dispose() {
    if (this.mesh) {
      this.mesh.geometry?.dispose();
      this.mesh.material?.dispose();
      this.mesh = null;
    }
  }
}

export default Chunk;
