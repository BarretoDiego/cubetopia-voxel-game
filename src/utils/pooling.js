/**
 * Sistema de pool de objetos para evitar alocações frequentes
 */

export class ObjectPool {
  constructor(factory, initialSize = 100, maxSize = 1000) {
    this.factory = factory;
    this.maxSize = maxSize;
    this.pool = [];

    // Pré-aloca objetos
    for (let i = 0; i < initialSize; i++) {
      this.pool.push(this.factory());
    }
  }

  acquire() {
    if (this.pool.length > 0) {
      return this.pool.pop();
    }
    return this.factory();
  }

  release(obj) {
    if (this.pool.length < this.maxSize) {
      // Reset do objeto se tiver método reset
      if (typeof obj.reset === "function") {
        obj.reset();
      }
      this.pool.push(obj);
    }
  }

  releaseAll(objects) {
    objects.forEach((obj) => this.release(obj));
  }

  get size() {
    return this.pool.length;
  }

  clear() {
    this.pool = [];
  }
}

/**
 * Pool específico para vetores 3D
 */
export class Vector3Pool extends ObjectPool {
  constructor(initialSize = 100) {
    super(() => ({ x: 0, y: 0, z: 0 }), initialSize);
  }

  acquire(x = 0, y = 0, z = 0) {
    const vec = super.acquire();
    vec.x = x;
    vec.y = y;
    vec.z = z;
    return vec;
  }
}

/**
 * Cache LRU (Least Recently Used) para chunks e outros dados
 */
export class LRUCache {
  constructor(maxSize = 100, onEvict = null) {
    this.maxSize = maxSize;
    this.cache = new Map();
    this.onEvict = onEvict; // Callback when item is evicted
  }

  get(key) {
    if (!this.cache.has(key)) return undefined;

    // Move para o final (mais recente)
    const value = this.cache.get(key);
    this.cache.delete(key);
    this.cache.set(key, value);
    return value;
  }

  set(key, value) {
    if (this.cache.has(key)) {
      this.cache.delete(key);
    } else if (this.cache.size >= this.maxSize) {
      // Remove o mais antigo (primeiro) and call eviction callback
      const firstKey = this.cache.keys().next().value;
      const evictedValue = this.cache.get(firstKey);
      this.cache.delete(firstKey);
      if (this.onEvict && evictedValue) {
        this.onEvict(evictedValue);
      }
    }
    this.cache.set(key, value);
  }

  has(key) {
    return this.cache.has(key);
  }

  delete(key) {
    return this.cache.delete(key);
  }

  clear() {
    this.cache.clear();
  }

  get size() {
    return this.cache.size;
  }

  keys() {
    return this.cache.keys();
  }

  values() {
    return this.cache.values();
  }

  entries() {
    return this.cache.entries();
  }
}

/**
 * Buffer de trabalho reutilizável para operações de mesh
 */
export class WorkBuffer {
  constructor(initialCapacity = 65536) {
    this.positions = new Float32Array(initialCapacity * 3);
    this.normals = new Float32Array(initialCapacity * 3);
    this.uvs = new Float32Array(initialCapacity * 2);
    this.colors = new Float32Array(initialCapacity * 3);
    this.indices = new Uint32Array(initialCapacity);
    this.vertexCount = 0;
    this.indexCount = 0;
  }

  ensureCapacity(requiredVertices, requiredIndices) {
    if (requiredVertices * 3 > this.positions.length) {
      const newSize = Math.max(requiredVertices * 3, this.positions.length * 2);
      this.positions = new Float32Array(newSize);
      this.normals = new Float32Array(newSize);
      this.colors = new Float32Array(newSize);
      this.uvs = new Float32Array((newSize / 3) * 2);
    }
    if (requiredIndices > this.indices.length) {
      this.indices = new Uint32Array(
        Math.max(requiredIndices, this.indices.length * 2)
      );
    }
  }

  reset() {
    this.vertexCount = 0;
    this.indexCount = 0;
  }
}
