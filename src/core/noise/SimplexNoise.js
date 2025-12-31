/**
 * Simplex Noise 2D/3D Implementation
 * Baseado no algoritmo de Ken Perlin e Stefan Gustavson
 */

export class SimplexNoise {
  constructor(seed = Math.random() * 10000) {
    this.seed = seed;

    // Constantes para 2D
    this.F2 = 0.5 * (Math.sqrt(3.0) - 1.0);
    this.G2 = (3.0 - Math.sqrt(3.0)) / 6.0;

    // Constantes para 3D
    this.F3 = 1.0 / 3.0;
    this.G3 = 1.0 / 6.0;

    // Tabela de permutação (512 para evitar overflow)
    this.perm = new Uint8Array(512);
    this.permMod12 = new Uint8Array(512);

    // Gradientes para 2D e 3D
    this.grad3 = new Float32Array([
      1, 1, 0, -1, 1, 0, 1, -1, 0, -1, -1, 0, 1, 0, 1, -1, 0, 1, 1, 0, -1, -1,
      0, -1, 0, 1, 1, 0, -1, 1, 0, 1, -1, 0, -1, -1,
    ]);

    this._initPermutation();
  }

  _initPermutation() {
    const p = new Uint8Array(256);

    // Gera permutação baseada no seed
    for (let i = 0; i < 256; i++) {
      p[i] = i;
    }

    // Fisher-Yates shuffle com seed
    let seed = this.seed;
    for (let i = 255; i > 0; i--) {
      seed = (seed * 16807) % 2147483647;
      const j = seed % (i + 1);
      [p[i], p[j]] = [p[j], p[i]];
    }

    // Duplica para evitar overflow
    for (let i = 0; i < 512; i++) {
      this.perm[i] = p[i & 255];
      this.permMod12[i] = this.perm[i] % 12;
    }
  }

  /**
   * Noise 2D
   */
  noise2D(xin, yin) {
    let n0, n1, n2;

    // Skew input space
    const s = (xin + yin) * this.F2;
    const i = Math.floor(xin + s);
    const j = Math.floor(yin + s);

    // Unskew back
    const t = (i + j) * this.G2;
    const X0 = i - t;
    const Y0 = j - t;
    const x0 = xin - X0;
    const y0 = yin - Y0;

    // Determina qual simplex
    let i1, j1;
    if (x0 > y0) {
      i1 = 1;
      j1 = 0;
    } else {
      i1 = 0;
      j1 = 1;
    }

    const x1 = x0 - i1 + this.G2;
    const y1 = y0 - j1 + this.G2;
    const x2 = x0 - 1.0 + 2.0 * this.G2;
    const y2 = y0 - 1.0 + 2.0 * this.G2;

    // Hash das coordenadas
    const ii = i & 255;
    const jj = j & 255;
    const gi0 = this.permMod12[ii + this.perm[jj]];
    const gi1 = this.permMod12[ii + i1 + this.perm[jj + j1]];
    const gi2 = this.permMod12[ii + 1 + this.perm[jj + 1]];

    // Calcula contribuição de cada canto
    let t0 = 0.5 - x0 * x0 - y0 * y0;
    if (t0 < 0) {
      n0 = 0.0;
    } else {
      t0 *= t0;
      n0 = t0 * t0 * (this.grad3[gi0 * 3] * x0 + this.grad3[gi0 * 3 + 1] * y0);
    }

    let t1 = 0.5 - x1 * x1 - y1 * y1;
    if (t1 < 0) {
      n1 = 0.0;
    } else {
      t1 *= t1;
      n1 = t1 * t1 * (this.grad3[gi1 * 3] * x1 + this.grad3[gi1 * 3 + 1] * y1);
    }

    let t2 = 0.5 - x2 * x2 - y2 * y2;
    if (t2 < 0) {
      n2 = 0.0;
    } else {
      t2 *= t2;
      n2 = t2 * t2 * (this.grad3[gi2 * 3] * x2 + this.grad3[gi2 * 3 + 1] * y2);
    }

    // Retorna valor no range [-1, 1]
    return 70.0 * (n0 + n1 + n2);
  }

  /**
   * Noise 3D
   */
  noise3D(xin, yin, zin) {
    let n0, n1, n2, n3;

    // Skew input space
    const s = (xin + yin + zin) * this.F3;
    const i = Math.floor(xin + s);
    const j = Math.floor(yin + s);
    const k = Math.floor(zin + s);

    // Unskew back
    const t = (i + j + k) * this.G3;
    const X0 = i - t;
    const Y0 = j - t;
    const Z0 = k - t;
    const x0 = xin - X0;
    const y0 = yin - Y0;
    const z0 = zin - Z0;

    // Determina qual simplex
    let i1, j1, k1, i2, j2, k2;

    if (x0 >= y0) {
      if (y0 >= z0) {
        i1 = 1;
        j1 = 0;
        k1 = 0;
        i2 = 1;
        j2 = 1;
        k2 = 0;
      } else if (x0 >= z0) {
        i1 = 1;
        j1 = 0;
        k1 = 0;
        i2 = 1;
        j2 = 0;
        k2 = 1;
      } else {
        i1 = 0;
        j1 = 0;
        k1 = 1;
        i2 = 1;
        j2 = 0;
        k2 = 1;
      }
    } else {
      if (y0 < z0) {
        i1 = 0;
        j1 = 0;
        k1 = 1;
        i2 = 0;
        j2 = 1;
        k2 = 1;
      } else if (x0 < z0) {
        i1 = 0;
        j1 = 1;
        k1 = 0;
        i2 = 0;
        j2 = 1;
        k2 = 1;
      } else {
        i1 = 0;
        j1 = 1;
        k1 = 0;
        i2 = 1;
        j2 = 1;
        k2 = 0;
      }
    }

    const x1 = x0 - i1 + this.G3;
    const y1 = y0 - j1 + this.G3;
    const z1 = z0 - k1 + this.G3;
    const x2 = x0 - i2 + 2.0 * this.G3;
    const y2 = y0 - j2 + 2.0 * this.G3;
    const z2 = z0 - k2 + 2.0 * this.G3;
    const x3 = x0 - 1.0 + 3.0 * this.G3;
    const y3 = y0 - 1.0 + 3.0 * this.G3;
    const z3 = z0 - 1.0 + 3.0 * this.G3;

    // Hash
    const ii = i & 255;
    const jj = j & 255;
    const kk = k & 255;
    const gi0 = this.permMod12[ii + this.perm[jj + this.perm[kk]]];
    const gi1 =
      this.permMod12[ii + i1 + this.perm[jj + j1 + this.perm[kk + k1]]];
    const gi2 =
      this.permMod12[ii + i2 + this.perm[jj + j2 + this.perm[kk + k2]]];
    const gi3 = this.permMod12[ii + 1 + this.perm[jj + 1 + this.perm[kk + 1]]];

    // Contribuições
    let t0 = 0.6 - x0 * x0 - y0 * y0 - z0 * z0;
    if (t0 < 0) {
      n0 = 0.0;
    } else {
      t0 *= t0;
      n0 =
        t0 *
        t0 *
        (this.grad3[gi0 * 3] * x0 +
          this.grad3[gi0 * 3 + 1] * y0 +
          this.grad3[gi0 * 3 + 2] * z0);
    }

    let t1 = 0.6 - x1 * x1 - y1 * y1 - z1 * z1;
    if (t1 < 0) {
      n1 = 0.0;
    } else {
      t1 *= t1;
      n1 =
        t1 *
        t1 *
        (this.grad3[gi1 * 3] * x1 +
          this.grad3[gi1 * 3 + 1] * y1 +
          this.grad3[gi1 * 3 + 2] * z1);
    }

    let t2 = 0.6 - x2 * x2 - y2 * y2 - z2 * z2;
    if (t2 < 0) {
      n2 = 0.0;
    } else {
      t2 *= t2;
      n2 =
        t2 *
        t2 *
        (this.grad3[gi2 * 3] * x2 +
          this.grad3[gi2 * 3 + 1] * y2 +
          this.grad3[gi2 * 3 + 2] * z2);
    }

    let t3 = 0.6 - x3 * x3 - y3 * y3 - z3 * z3;
    if (t3 < 0) {
      n3 = 0.0;
    } else {
      t3 *= t3;
      n3 =
        t3 *
        t3 *
        (this.grad3[gi3 * 3] * x3 +
          this.grad3[gi3 * 3 + 1] * y3 +
          this.grad3[gi3 * 3 + 2] * z3);
    }

    return 32.0 * (n0 + n1 + n2 + n3);
  }
}

export default SimplexNoise;
