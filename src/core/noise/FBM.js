/**
 * Fractal Brownian Motion (FBM)
 * Combina múltiplas camadas de noise para terreno mais natural
 */

export class FBM {
  constructor(options = {}) {
    this.octaves = options.octaves || 6;
    this.lacunarity = options.lacunarity || 2.0; // Frequência multiplier
    this.persistence = options.persistence || 0.5; // Amplitude multiplier
    this.scale = options.scale || 1.0;
    this.offsetX = options.offsetX || 0;
    this.offsetZ = options.offsetZ || 0;
  }

  /**
   * Amostra FBM 2D
   * @param {SimplexNoise} noise - Instância de noise
   * @param {number} x - Coordenada X
   * @param {number} z - Coordenada Z
   * @returns {number} Valor no range aproximado [-1, 1]
   */
  sample2D(noise, x, z) {
    let value = 0;
    let amplitude = 1;
    let frequency = this.scale;
    let maxValue = 0;

    for (let i = 0; i < this.octaves; i++) {
      value +=
        amplitude *
        noise.noise2D(
          (x + this.offsetX) * frequency,
          (z + this.offsetZ) * frequency
        );
      maxValue += amplitude;
      amplitude *= this.persistence;
      frequency *= this.lacunarity;
    }

    return value / maxValue;
  }

  /**
   * Amostra FBM 3D
   * @param {SimplexNoise} noise - Instância de noise
   * @param {number} x - Coordenada X
   * @param {number} y - Coordenada Y
   * @param {number} z - Coordenada Z
   * @returns {number} Valor no range aproximado [-1, 1]
   */
  sample3D(noise, x, y, z) {
    let value = 0;
    let amplitude = 1;
    let frequency = this.scale;
    let maxValue = 0;

    for (let i = 0; i < this.octaves; i++) {
      value +=
        amplitude *
        noise.noise3D(
          (x + this.offsetX) * frequency,
          y * frequency,
          (z + this.offsetZ) * frequency
        );
      maxValue += amplitude;
      amplitude *= this.persistence;
      frequency *= this.lacunarity;
    }

    return value / maxValue;
  }

  /**
   * FBM com ridged noise (para montanhas)
   */
  ridged2D(noise, x, z) {
    let value = 0;
    let amplitude = 1;
    let frequency = this.scale;
    let maxValue = 0;

    for (let i = 0; i < this.octaves; i++) {
      let n = noise.noise2D(
        (x + this.offsetX) * frequency,
        (z + this.offsetZ) * frequency
      );
      n = 1 - Math.abs(n); // Ridge
      n = n * n; // Sharpen
      value += amplitude * n;
      maxValue += amplitude;
      amplitude *= this.persistence;
      frequency *= this.lacunarity;
    }

    return value / maxValue;
  }

  /**
   * FBM turbulento (para nuvens, erosão)
   */
  turbulence2D(noise, x, z) {
    let value = 0;
    let amplitude = 1;
    let frequency = this.scale;
    let maxValue = 0;

    for (let i = 0; i < this.octaves; i++) {
      value +=
        amplitude *
        Math.abs(
          noise.noise2D(
            (x + this.offsetX) * frequency,
            (z + this.offsetZ) * frequency
          )
        );
      maxValue += amplitude;
      amplitude *= this.persistence;
      frequency *= this.lacunarity;
    }

    return value / maxValue;
  }

  /**
   * Domain warping para terreno mais interessante
   */
  warped2D(noise, x, z, warpAmount = 20) {
    const warpX = this.sample2D(noise, x * 0.5, z * 0.5) * warpAmount;
    const warpZ =
      this.sample2D(noise, x * 0.5 + 100, z * 0.5 + 100) * warpAmount;
    return this.sample2D(noise, x + warpX, z + warpZ);
  }
}

export default FBM;
