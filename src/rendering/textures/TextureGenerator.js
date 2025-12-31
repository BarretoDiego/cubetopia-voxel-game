/**
 * Gerador de Texturas Procedurais
 */

import * as THREE from "three";

export class TextureGenerator {
  constructor() {
    this.cache = new Map();
    this.canvas = null;
    this.ctx = null;
    this.size = 64;
  }

  _initCanvas() {
    if (!this.canvas && typeof document !== "undefined") {
      this.canvas = document.createElement("canvas");
      this.canvas.width = this.size;
      this.canvas.height = this.size;
      this.ctx = this.canvas.getContext("2d");
    }
  }

  /**
   * Cria textura procedural
   */
  create(options) {
    const cacheKey = JSON.stringify(options);
    if (this.cache.has(cacheKey)) {
      return this.cache.get(cacheKey);
    }

    this._initCanvas();
    if (!this.ctx) return null;

    const {
      baseColor = "#808080",
      noise = { amount: 0.1, scale: 1 },
      pattern = null,
      border = false,
      gradient = null,
      overlay = null,
    } = options;

    // Limpa canvas
    this.ctx.clearRect(0, 0, this.size, this.size);

    // Cor base ou gradiente
    if (gradient) {
      this._applyGradient(gradient);
    } else {
      this.ctx.fillStyle = baseColor;
      this.ctx.fillRect(0, 0, this.size, this.size);
    }

    // Aplica noise com variação de cor para simular textura
    if (noise && noise.amount > 0) {
      this._applyNoise(noise.amount, noise.scale);
    }

    // Add shading noise (simula normal map rudimentar)
    this._applyShadingNoise(0.05);

    // Aplica padrão
    if (pattern) {
      this._applyPattern(pattern, baseColor);
    }

    // Aplica overlay
    if (overlay) {
      this._applyOverlay(overlay, baseColor);
    }

    // Borda
    if (border) {
      this._applyBorder();
    }

    // Cria textura Three.js
    const texture = new THREE.CanvasTexture(this.canvas);
    texture.magFilter = THREE.NearestFilter;
    texture.minFilter = THREE.NearestFilter;
    texture.needsUpdate = true;

    this.cache.set(cacheKey, texture);
    return texture;
  }

  /**
   * Aplica gradiente
   */
  _applyGradient(gradient) {
    const { direction = "vertical", colors } = gradient;
    let grad;

    if (direction === "vertical") {
      grad = this.ctx.createLinearGradient(0, 0, 0, this.size);
    } else if (direction === "horizontal") {
      grad = this.ctx.createLinearGradient(0, 0, this.size, 0);
    } else {
      grad = this.ctx.createRadialGradient(
        this.size / 2,
        this.size / 2,
        0,
        this.size / 2,
        this.size / 2,
        this.size / 2
      );
    }

    colors.forEach((color, i) => {
      grad.addColorStop(i / (colors.length - 1), color);
    });

    this.ctx.fillStyle = grad;
    this.ctx.fillRect(0, 0, this.size, this.size);
  }

  /**
   * Aplica noise
   */
  _applyNoise(amount, scale) {
    const imageData = this.ctx.getImageData(0, 0, this.size, this.size);
    const data = imageData.data;

    for (let i = 0; i < data.length; i += 4) {
      const noise = (Math.random() - 0.5) * 255 * amount;
      data[i] = Math.max(0, Math.min(255, data[i] + noise));
      data[i + 1] = Math.max(0, Math.min(255, data[i + 1] + noise));
      data[i + 2] = Math.max(0, Math.min(255, data[i + 2] + noise));
    }

    this.ctx.putImageData(imageData, 0, 0);
  }

  /**
   * Aplica padrão
   */
  _applyPattern(pattern, baseColor) {
    const rgb = this._hexToRgb(baseColor);

    switch (pattern) {
      case "brick":
        this._drawBrickPattern(rgb);
        break;
      case "wood":
        this._drawWoodPattern(rgb);
        break;
      case "stone":
        this._drawStonePattern(rgb);
        break;
      case "leaves":
        this._drawLeavesPattern(rgb);
        break;
      case "ore":
        this._drawOrePattern(rgb);
        break;
    }
  }

  _drawBrickPattern(rgb) {
    this.ctx.strokeStyle = `rgba(0,0,0,0.3)`;
    this.ctx.lineWidth = 2;

    const brickHeight = 16;
    const brickWidth = 32;

    for (let y = 0; y < this.size; y += brickHeight) {
      const offset = (Math.floor(y / brickHeight) % 2) * (brickWidth / 2);
      for (let x = -brickWidth + offset; x < this.size; x += brickWidth) {
        this.ctx.strokeRect(x, y, brickWidth, brickHeight);
      }
    }
  }

  _drawWoodPattern(rgb) {
    this.ctx.strokeStyle = `rgba(0,0,0,0.2)`;
    this.ctx.lineWidth = 1;

    for (let y = 0; y < this.size; y += 4 + Math.random() * 4) {
      this.ctx.beginPath();
      this.ctx.moveTo(0, y);
      this.ctx.lineTo(this.size, y + (Math.random() - 0.5) * 2);
      this.ctx.stroke();
    }
  }

  _drawStonePattern(rgb) {
    this.ctx.fillStyle = `rgba(0,0,0,0.1)`;

    for (let i = 0; i < 15; i++) {
      const x = Math.random() * this.size;
      const y = Math.random() * this.size;
      const r = 2 + Math.random() * 6;
      this.ctx.beginPath();
      this.ctx.arc(x, y, r, 0, Math.PI * 2);
      this.ctx.fill();
    }
  }

  _drawLeavesPattern(rgb) {
    // Pontos escuros simulando profundidade
    this.ctx.fillStyle = `rgba(0,0,0,0.3)`;
    for (let i = 0; i < 20; i++) {
      const x = Math.random() * this.size;
      const y = Math.random() * this.size;
      this.ctx.fillRect(x, y, 3, 3);
    }

    // Pontos claros simulando luz
    this.ctx.fillStyle = `rgba(255,255,255,0.2)`;
    for (let i = 0; i < 10; i++) {
      const x = Math.random() * this.size;
      const y = Math.random() * this.size;
      this.ctx.fillRect(x, y, 2, 2);
    }
  }

  _drawOrePattern(rgb) {
    // Veios de minério
    const oreColors = {
      coal: "#1a1a1a",
      iron: "#d4a574",
      gold: "#ffd700",
      diamond: "#00ffff",
    };

    this.ctx.fillStyle = oreColors[rgb.ore] || "#ffffff";

    for (let i = 0; i < 5; i++) {
      const x = 10 + Math.random() * (this.size - 20);
      const y = 10 + Math.random() * (this.size - 20);

      // Cluster de pixels
      for (let j = 0; j < 4; j++) {
        const ox = x + (Math.random() - 0.5) * 8;
        const oy = y + (Math.random() - 0.5) * 8;
        this.ctx.fillRect(ox, oy, 3, 3);
      }
    }
  }

  /**
   * Aplica overlay
   */
  _applyOverlay(overlay, baseColor) {
    switch (overlay) {
      case "grass_blades":
        this._drawGrassBlades();
        break;
      case "dirt_top":
        this._drawDirtTop();
        break;
    }
  }

  _drawGrassBlades() {
    this.ctx.strokeStyle = "rgba(34, 139, 34, 0.5)";
    this.ctx.lineWidth = 1;

    for (let i = 0; i < 30; i++) {
      const x = Math.random() * this.size;
      const y = this.size - Math.random() * 10;
      const height = 3 + Math.random() * 8;

      this.ctx.beginPath();
      this.ctx.moveTo(x, this.size);
      this.ctx.lineTo(x + (Math.random() - 0.5) * 3, y - height);
      this.ctx.stroke();
    }
  }

  _drawDirtTop() {
    this.ctx.fillStyle = "rgba(139, 69, 19, 0.5)";

    for (let x = 0; x < this.size; x += 2) {
      const height = 4 + Math.random() * 8;
      this.ctx.fillRect(x, this.size - height, 2, height);
    }
  }

  /**
   * Aplica borda
   */
  _applyBorder() {
    this.ctx.strokeStyle = "rgba(0,0,0,0.15)";
    this.ctx.lineWidth = 1;
    this.ctx.strokeRect(0, 0, this.size, this.size);
  }

  /**
   * Aplica shading noise
   */
  _applyShadingNoise(amount) {
    const imageData = this.ctx.getImageData(0, 0, this.size, this.size);
    const data = imageData.data;

    for (let i = 0; i < data.length; i += 4) {
      // Varia levemente cada componente para dar "textura"
      const shade = (Math.random() - 0.5) * 255 * amount;
      data[i] = Math.max(0, Math.min(255, data[i] + shade));
      data[i + 1] = Math.max(0, Math.min(255, data[i + 1] + shade));
      data[i + 2] = Math.max(0, Math.min(255, data[i + 2] + shade));
    }

    this.ctx.putImageData(imageData, 0, 0);
  }

  /**
   * Converte hex para RGB
   */
  _hexToRgb(hex) {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
    return result
      ? {
          r: parseInt(result[1], 16),
          g: parseInt(result[2], 16),
          b: parseInt(result[3], 16),
        }
      : { r: 128, g: 128, b: 128 };
  }

  /**
   * Limpa cache
   */
  clearCache() {
    this.cache.forEach((texture) => texture.dispose());
    this.cache.clear();
  }
}

// Singleton
export const textureGenerator = new TextureGenerator();
export default textureGenerator;
