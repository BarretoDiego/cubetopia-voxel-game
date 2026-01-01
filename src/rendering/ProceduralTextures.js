/**
 * Procedural Texture Generator
 * Creates canvas-based textures for each block type
 */

/**
 * Generate a procedural texture for a block type
 * @param {string} color - Base color in hex
 * @param {string} type - Block material type
 * @param {number} size - Texture size (default 16x16)
 * @returns {HTMLCanvasElement}
 */
export function generateBlockTexture(color, type = "solid", size = 16) {
  const canvas = document.createElement("canvas");
  canvas.width = size;
  canvas.height = size;
  const ctx = canvas.getContext("2d");

  // Parse color
  const rgb = hexToRgb(color);

  // Fill base color
  ctx.fillStyle = color;
  ctx.fillRect(0, 0, size, size);

  // Add noise/pattern based on type
  switch (type) {
    case "dirt":
      addDirtNoise(ctx, rgb, size);
      break;
    case "stone":
      addStoneNoise(ctx, rgb, size);
      break;
    case "wood":
      addWoodGrain(ctx, rgb, size);
      break;
    case "grass":
      addGrassPattern(ctx, rgb, size);
      break;
    case "sand":
      addSandNoise(ctx, rgb, size);
      break;
    case "water":
      addWaterPattern(ctx, rgb, size);
      break;
    case "brick":
      addBrickPattern(ctx, rgb, size);
      break;
    case "ore":
      addOrePattern(ctx, rgb, size);
      break;
    default:
      addGenericNoise(ctx, rgb, size);
  }

  return canvas;
}

function hexToRgb(hex) {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result
    ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16),
      }
    : { r: 128, g: 128, b: 128 };
}

function addDirtNoise(ctx, rgb, size) {
  for (let i = 0; i < (size * size) / 4; i++) {
    const x = Math.random() * size;
    const y = Math.random() * size;
    const shade = Math.random() * 30 - 15;
    ctx.fillStyle = `rgb(${rgb.r + shade}, ${rgb.g + shade}, ${rgb.b + shade})`;
    ctx.fillRect(x, y, 2, 2);
  }
}

function addStoneNoise(ctx, rgb, size) {
  // Add cracks and variations
  for (let i = 0; i < (size * size) / 3; i++) {
    const x = Math.random() * size;
    const y = Math.random() * size;
    const shade = Math.random() * 40 - 20;
    ctx.fillStyle = `rgb(${rgb.r + shade}, ${rgb.g + shade}, ${rgb.b + shade})`;
    ctx.fillRect(x, y, 1, 1);
  }
  // Dark cracks
  ctx.strokeStyle = `rgb(${rgb.r - 30}, ${rgb.g - 30}, ${rgb.b - 30})`;
  for (let i = 0; i < 3; i++) {
    ctx.beginPath();
    ctx.moveTo(Math.random() * size, Math.random() * size);
    ctx.lineTo(Math.random() * size, Math.random() * size);
    ctx.stroke();
  }
}

function addWoodGrain(ctx, rgb, size) {
  // Horizontal wood grain lines
  for (let y = 0; y < size; y += 2) {
    const shade = y % 4 === 0 ? -20 : 10;
    ctx.fillStyle = `rgb(${rgb.r + shade}, ${rgb.g + shade}, ${rgb.b + shade})`;
    ctx.fillRect(0, y, size, 1);
  }
  // Add knots
  for (let i = 0; i < 2; i++) {
    const x = Math.random() * (size - 4) + 2;
    const y = Math.random() * (size - 4) + 2;
    ctx.fillStyle = `rgb(${rgb.r - 40}, ${rgb.g - 40}, ${rgb.b - 40})`;
    ctx.beginPath();
    ctx.arc(x, y, 2, 0, Math.PI * 2);
    ctx.fill();
  }
}

function addGrassPattern(ctx, rgb, size) {
  // Green grass blades on top
  for (let x = 0; x < size; x++) {
    const h = Math.random() * 4 + 1;
    const shade = Math.random() * 30;
    ctx.fillStyle = `rgb(${rgb.r + shade}, ${rgb.g + shade}, ${rgb.b})`;
    ctx.fillRect(x, 0, 1, h);
  }
}

function addSandNoise(ctx, rgb, size) {
  for (let i = 0; i < (size * size) / 2; i++) {
    const x = Math.random() * size;
    const y = Math.random() * size;
    const shade = Math.random() * 20 - 10;
    ctx.fillStyle = `rgb(${rgb.r + shade}, ${rgb.g + shade}, ${rgb.b + shade})`;
    ctx.fillRect(x, y, 1, 1);
  }
}

function addWaterPattern(ctx, rgb, size) {
  // Wave-like lighter areas
  for (let y = 0; y < size; y += 4) {
    const offset = Math.sin(y * 0.5) * 3;
    ctx.fillStyle = `rgba(255, 255, 255, 0.2)`;
    ctx.fillRect(offset, y, size, 2);
  }
}

function addBrickPattern(ctx, rgb, size) {
  // Brick pattern with mortar
  const mortarColor = `rgb(${rgb.r + 40}, ${rgb.g + 40}, ${rgb.b + 40})`;
  ctx.strokeStyle = mortarColor;
  ctx.lineWidth = 1;

  // Horizontal lines
  for (let y = 4; y < size; y += 4) {
    ctx.beginPath();
    ctx.moveTo(0, y);
    ctx.lineTo(size, y);
    ctx.stroke();
  }
  // Vertical lines (staggered)
  for (let y = 0; y < size; y += 4) {
    const offset = (Math.floor(y / 4) % 2) * (size / 2);
    ctx.beginPath();
    ctx.moveTo(offset, y);
    ctx.lineTo(offset, y + 4);
    ctx.moveTo((offset + size / 2) % size, y);
    ctx.lineTo((offset + size / 2) % size, y + 4);
    ctx.stroke();
  }
}

function addOrePattern(ctx, rgb, size) {
  // Add shiny ore spots
  for (let i = 0; i < 5; i++) {
    const x = Math.random() * (size - 3) + 1;
    const y = Math.random() * (size - 3) + 1;
    ctx.fillStyle = `rgb(${Math.min(255, rgb.r + 80)}, ${Math.min(
      255,
      rgb.g + 80
    )}, ${Math.min(255, rgb.b + 80)})`;
    ctx.fillRect(x, y, 2, 2);
  }
}

function addGenericNoise(ctx, rgb, size) {
  for (let i = 0; i < (size * size) / 6; i++) {
    const x = Math.random() * size;
    const y = Math.random() * size;
    const shade = Math.random() * 20 - 10;
    ctx.fillStyle = `rgb(${rgb.r + shade}, ${rgb.g + shade}, ${rgb.b + shade})`;
    ctx.fillRect(x, y, 1, 1);
  }
}

/**
 * Generate all block textures as a texture atlas or map
 */
export function generateAllTextures(blockDefinitions) {
  const textures = {};

  for (const [id, def] of Object.entries(blockDefinitions)) {
    if (def.color) {
      let type = "solid";
      if (def.name?.includes("Terra") || def.name?.includes("Dirt"))
        type = "dirt";
      else if (def.name?.includes("Pedra") || def.name?.includes("Stone"))
        type = "stone";
      else if (
        def.name?.includes("Madeira") ||
        def.name?.includes("Wood") ||
        def.name?.includes("Log")
      )
        type = "wood";
      else if (def.name?.includes("Grama") || def.name?.includes("Grass"))
        type = "grass";
      else if (def.name?.includes("Areia") || def.name?.includes("Sand"))
        type = "sand";
      else if (def.name?.includes("Água") || def.name?.includes("Water"))
        type = "water";
      else if (def.name?.includes("Tijolo") || def.name?.includes("Brick"))
        type = "brick";
      else if (def.name?.includes("Ore") || def.name?.includes("Minério"))
        type = "ore";

      textures[id] = generateBlockTexture(def.color, type);
    }
  }

  return textures;
}

export default { generateBlockTexture, generateAllTextures };
