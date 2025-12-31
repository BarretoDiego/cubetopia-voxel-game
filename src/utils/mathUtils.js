/**
 * Utilitários matemáticos para geração procedural
 */

/**
 * Gera um número pseudo-aleatório baseado em seed
 */
export function seededRandom(seed) {
  const x = Math.sin(seed) * 10000;
  return x - Math.floor(x);
}

/**
 * Interpolação linear
 */
export function lerp(a, b, t) {
  return a + (b - a) * t;
}

/**
 * Interpolação suave (smoothstep)
 */
export function smoothstep(edge0, edge1, x) {
  const t = Math.max(0, Math.min(1, (x - edge0) / (edge1 - edge0)));
  return t * t * (3 - 2 * t);
}

/**
 * Interpolação ainda mais suave (smootherstep)
 */
export function smootherstep(edge0, edge1, x) {
  const t = Math.max(0, Math.min(1, (x - edge0) / (edge1 - edge0)));
  return t * t * t * (t * (t * 6 - 15) + 10);
}

/**
 * Clamp - limita valor entre min e max
 */
export function clamp(value, min, max) {
  return Math.max(min, Math.min(max, value));
}

/**
 * Mapeia valor de uma faixa para outra
 */
export function map(value, inMin, inMax, outMin, outMax) {
  return outMin + ((value - inMin) / (inMax - inMin)) * (outMax - outMin);
}

/**
 * Modulo que funciona corretamente com negativos
 */
export function mod(n, m) {
  return ((n % m) + m) % m;
}

/**
 * Converte coordenadas do mundo para coordenadas de chunk
 */
export function worldToChunk(x, z, chunkSize) {
  return {
    cx: Math.floor(x / chunkSize),
    cz: Math.floor(z / chunkSize),
  };
}

/**
 * Converte coordenadas do mundo para coordenadas locais do chunk
 */
export function worldToLocal(x, y, z, chunkSize) {
  return {
    lx: mod(x, chunkSize),
    ly: y,
    lz: mod(z, chunkSize),
  };
}

/**
 * Converte coordenadas locais + chunk para índice no array
 */
export function coordsToIndex(lx, ly, lz, chunkSize, chunkHeight) {
  return lx + ly * chunkSize + lz * chunkSize * chunkHeight;
}

/**
 * Converte índice para coordenadas
 */
export function indexToCoords(index, chunkSize, chunkHeight) {
  const lx = index % chunkSize;
  const ly = Math.floor(index / chunkSize) % chunkHeight;
  const lz = Math.floor(index / (chunkSize * chunkHeight));
  return { lx, ly, lz };
}

/**
 * Distância euclidiana 2D
 */
export function distance2D(x1, z1, x2, z2) {
  const dx = x2 - x1;
  const dz = z2 - z1;
  return Math.sqrt(dx * dx + dz * dz);
}

/**
 * Distância euclidiana 3D
 */
export function distance3D(x1, y1, z1, x2, y2, z2) {
  const dx = x2 - x1;
  const dy = y2 - y1;
  const dz = z2 - z1;
  return Math.sqrt(dx * dx + dy * dy + dz * dz);
}

/**
 * Distância de Manhattan 3D
 */
export function manhattanDistance(x1, y1, z1, x2, y2, z2) {
  return Math.abs(x2 - x1) + Math.abs(y2 - y1) + Math.abs(z2 - z1);
}

/**
 * Hash simples para coordenadas (para uso em seeds)
 */
export function hashCoords(x, y, z) {
  let hash = 17;
  hash = hash * 31 + (x | 0);
  hash = hash * 31 + (y | 0);
  hash = hash * 31 + (z | 0);
  return hash;
}

/**
 * Gerador Linear Congruencial (LCG) para números pseudo-aleatórios
 */
export class SeededRNG {
  constructor(seed = Date.now()) {
    this.seed = seed;
    this.m = 0x80000000; // 2^31
    this.a = 1103515245;
    this.c = 12345;
    this.state = seed;
  }

  next() {
    this.state = (this.a * this.state + this.c) % this.m;
    return this.state / this.m;
  }

  nextInt(min, max) {
    return Math.floor(this.next() * (max - min + 1)) + min;
  }

  nextFloat(min = 0, max = 1) {
    return min + this.next() * (max - min);
  }

  nextBool(probability = 0.5) {
    return this.next() < probability;
  }

  /**
   * Escolhe item aleatório de um array
   */
  choose(array) {
    return array[this.nextInt(0, array.length - 1)];
  }

  /**
   * Embaralha array (Fisher-Yates)
   */
  shuffle(array) {
    const result = [...array];
    for (let i = result.length - 1; i > 0; i--) {
      const j = this.nextInt(0, i);
      [result[i], result[j]] = [result[j], result[i]];
    }
    return result;
  }
}
