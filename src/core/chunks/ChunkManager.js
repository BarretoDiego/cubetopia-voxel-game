/**
 * Gerenciador de Chunks - Controla carregamento e descarregamento
 */

import { CHUNK_SIZE, RENDER_DISTANCE } from "../../utils/constants.js";
import { worldToChunk, distance2D } from "../../utils/mathUtils.js";
import { LRUCache } from "../../utils/pooling.js";
import { Chunk } from "./Chunk.js";

export class ChunkManager {
  constructor(worldGenerator) {
    this.worldGenerator = worldGenerator;
    this.chunks = new Map();
    // Cache com callback de eviction para limpar recursos
    this.chunkCache = new LRUCache(100, (chunk) => {
      if (chunk && chunk.dispose) {
        chunk.dispose();
      }
    });
    this.loadQueue = [];
    this.isLoading = false;

    // Listeners
    this.onChunkLoaded = null;
    this.onChunkUnloaded = null;
  }

  /**
   * Obtém ou cria um chunk
   */
  getChunk(cx, cz) {
    const id = `${cx},${cz}`;

    if (this.chunks.has(id)) {
      return this.chunks.get(id);
    }

    // Verifica cache
    if (this.chunkCache.has(id)) {
      const chunk = this.chunkCache.get(id);
      this.chunkCache.delete(id);
      this.chunks.set(id, chunk);
      return chunk;
    }

    return null;
  }

  /**
   * Carrega chunk de forma assíncrona
   */
  async loadChunk(cx, cz) {
    const id = `${cx},${cz}`;

    if (this.chunks.has(id)) {
      return this.chunks.get(id);
    }

    // Verifica cache
    if (this.chunkCache.has(id)) {
      const chunk = this.chunkCache.get(id);
      this.chunkCache.delete(id);
      this.chunks.set(id, chunk);
      return chunk;
    }

    // Cria e gera novo chunk
    const chunk = new Chunk(cx, cz);

    if (this.worldGenerator) {
      this.worldGenerator.generateChunk(chunk);
    }

    chunk.isGenerated = true;
    this.chunks.set(id, chunk);

    if (this.onChunkLoaded) {
      this.onChunkLoaded(chunk);
    }

    return chunk;
  }

  /**
   * Descarrega chunk para cache
   */
  unloadChunk(cx, cz) {
    const id = `${cx},${cz}`;
    const chunk = this.chunks.get(id);

    if (chunk) {
      this.chunkCache.set(id, chunk);
      this.chunks.delete(id);

      if (this.onChunkUnloaded) {
        this.onChunkUnloaded(chunk);
      }
    }
  }

  /**
   * Atualiza chunks baseado na posição do jogador
   */
  updateAroundPlayer(playerX, playerZ) {
    const { cx: playerCx, cz: playerCz } = worldToChunk(
      playerX,
      playerZ,
      CHUNK_SIZE
    );
    const chunksToLoad = [];
    const chunksToKeep = new Set();

    // Determina chunks que devem estar carregados
    for (let dx = -RENDER_DISTANCE; dx <= RENDER_DISTANCE; dx++) {
      for (let dz = -RENDER_DISTANCE; dz <= RENDER_DISTANCE; dz++) {
        const cx = playerCx + dx;
        const cz = playerCz + dz;
        const id = `${cx},${cz}`;

        chunksToKeep.add(id);

        if (!this.chunks.has(id)) {
          chunksToLoad.push({ cx, cz, dist: Math.abs(dx) + Math.abs(dz) });
        }
      }
    }

    // Descarrega chunks distantes
    for (const [id, chunk] of this.chunks) {
      if (!chunksToKeep.has(id)) {
        this.unloadChunk(chunk.cx, chunk.cz);
      }
    }

    // Ordena por distância (mais próximos primeiro)
    chunksToLoad.sort((a, b) => a.dist - b.dist);

    // Carrega chunks necessários
    return chunksToLoad.map(({ cx, cz }) => this.loadChunk(cx, cz));
  }

  /**
   * Obtém bloco em coordenadas do mundo
   */
  getBlock(wx, wy, wz) {
    const { cx, cz } = worldToChunk(wx, wz, CHUNK_SIZE);
    const chunk = this.getChunk(cx, cz);

    if (!chunk) return 0; // AIR

    const lx = ((wx % CHUNK_SIZE) + CHUNK_SIZE) % CHUNK_SIZE;
    const lz = ((wz % CHUNK_SIZE) + CHUNK_SIZE) % CHUNK_SIZE;

    return chunk.getBlock(lx, wy, lz);
  }

  /**
   * Define bloco em coordenadas do mundo
   */
  setBlock(wx, wy, wz, type) {
    const { cx, cz } = worldToChunk(wx, wz, CHUNK_SIZE);
    const chunk = this.getChunk(cx, cz);

    if (!chunk) return false;

    const lx = ((wx % CHUNK_SIZE) + CHUNK_SIZE) % CHUNK_SIZE;
    const lz = ((wz % CHUNK_SIZE) + CHUNK_SIZE) % CHUNK_SIZE;

    const result = chunk.setBlock(lx, wy, lz, type);

    // Marca chunks vizinhos como dirty se bloco estiver na borda
    if (result) {
      if (lx === 0) this._markDirty(cx - 1, cz);
      if (lx === CHUNK_SIZE - 1) this._markDirty(cx + 1, cz);
      if (lz === 0) this._markDirty(cx, cz - 1);
      if (lz === CHUNK_SIZE - 1) this._markDirty(cx, cz + 1);
    }

    return result;
  }

  /**
   * Marca chunk como dirty (precisa rebuild de mesh)
   */
  _markDirty(cx, cz) {
    const chunk = this.getChunk(cx, cz);
    if (chunk) {
      chunk.isDirty = true;
    }
  }

  /**
   * Obtém altura do terreno em uma posição
   */
  getHeight(wx, wz) {
    const { cx, cz } = worldToChunk(wx, wz, CHUNK_SIZE);
    const chunk = this.getChunk(cx, cz);

    if (!chunk) return 0;

    const lx = ((wx % CHUNK_SIZE) + CHUNK_SIZE) % CHUNK_SIZE;
    const lz = ((wz % CHUNK_SIZE) + CHUNK_SIZE) % CHUNK_SIZE;

    return chunk.getHeight(lx, lz);
  }

  /**
   * Obtém lista de chunks carregados
   */
  getLoadedChunks() {
    return Array.from(this.chunks.values());
  }

  /**
   * Obtém chunks que precisam de rebuild
   */
  getDirtyChunks() {
    return this.getLoadedChunks().filter((chunk) => chunk.isDirty);
  }

  /**
   * Número de chunks carregados
   */
  get loadedCount() {
    return this.chunks.size;
  }

  /**
   * Limpa todos os chunks
   */
  clear() {
    for (const chunk of this.chunks.values()) {
      chunk.dispose();
    }
    this.chunks.clear();
    this.chunkCache.clear();
  }
}

export default ChunkManager;
