/**
 * Presets de texturas para blocos
 */

import { BlockTypes } from "../../core/blocks/BlockTypes.js";
import { textureGenerator } from "./TextureGenerator.js";

export function createBlockTextures() {
  return {
    [BlockTypes.GRASS]: textureGenerator.create({
      baseColor: "#567d46",
      noise: { amount: 0.15, scale: 1 },
      pattern: "leaves",
      border: true,
    }),

    [BlockTypes.DIRT]: textureGenerator.create({
      baseColor: "#8b6914",
      noise: { amount: 0.2, scale: 1 },
      border: true,
    }),

    [BlockTypes.STONE]: textureGenerator.create({
      baseColor: "#7a7a7a",
      noise: { amount: 0.15, scale: 1 },
      pattern: "stone",
      border: true,
    }),

    [BlockTypes.WOOD]: textureGenerator.create({
      baseColor: "#8b5a2b",
      noise: { amount: 0.1, scale: 1 },
      pattern: "wood",
      border: true,
    }),

    [BlockTypes.LEAVES]: textureGenerator.create({
      baseColor: "#228b22",
      noise: { amount: 0.2, scale: 1 },
      pattern: "leaves",
    }),

    [BlockTypes.SAND]: textureGenerator.create({
      baseColor: "#e0c090",
      noise: { amount: 0.15, scale: 1 },
      border: true,
    }),

    [BlockTypes.WATER]: textureGenerator.create({
      baseColor: "#3498db",
      noise: { amount: 0.1, scale: 1 },
      gradient: {
        direction: "vertical",
        colors: ["#5dade2", "#2980b9"],
      },
    }),

    [BlockTypes.SNOW]: textureGenerator.create({
      baseColor: "#f0f0f0",
      noise: { amount: 0.05, scale: 1 },
      border: true,
    }),

    [BlockTypes.ICE]: textureGenerator.create({
      baseColor: "#a5f2f3",
      noise: { amount: 0.1, scale: 1 },
      gradient: {
        direction: "radial",
        colors: ["#c8f7f7", "#7fcdcd"],
      },
    }),

    [BlockTypes.COAL_ORE]: textureGenerator.create({
      baseColor: "#5a5a5a",
      noise: { amount: 0.1, scale: 1 },
      pattern: "ore",
    }),

    [BlockTypes.IRON_ORE]: textureGenerator.create({
      baseColor: "#7a7a7a",
      noise: { amount: 0.1, scale: 1 },
    }),

    [BlockTypes.GOLD_ORE]: textureGenerator.create({
      baseColor: "#7a7a7a",
      noise: { amount: 0.1, scale: 1 },
    }),

    [BlockTypes.DIAMOND_ORE]: textureGenerator.create({
      baseColor: "#7a7a7a",
      noise: { amount: 0.1, scale: 1 },
    }),

    [BlockTypes.OAK_LOG]: textureGenerator.create({
      baseColor: "#6b4423",
      noise: { amount: 0.1, scale: 1 },
      pattern: "wood",
      border: true,
    }),

    [BlockTypes.BIRCH_LOG]: textureGenerator.create({
      baseColor: "#d5c4a1",
      noise: { amount: 0.1, scale: 1 },
      pattern: "wood",
      border: true,
    }),

    [BlockTypes.SPRUCE_LOG]: textureGenerator.create({
      baseColor: "#3e2723",
      noise: { amount: 0.1, scale: 1 },
      pattern: "wood",
      border: true,
    }),

    [BlockTypes.OAK_LEAVES]: textureGenerator.create({
      baseColor: "#228b22",
      noise: { amount: 0.2, scale: 1 },
      pattern: "leaves",
    }),

    [BlockTypes.BIRCH_LEAVES]: textureGenerator.create({
      baseColor: "#80c622",
      noise: { amount: 0.2, scale: 1 },
      pattern: "leaves",
    }),

    [BlockTypes.SPRUCE_LEAVES]: textureGenerator.create({
      baseColor: "#1a472a",
      noise: { amount: 0.15, scale: 1 },
      pattern: "leaves",
    }),

    [BlockTypes.CACTUS]: textureGenerator.create({
      baseColor: "#0b5d1e",
      noise: { amount: 0.1, scale: 1 },
      border: true,
    }),

    [BlockTypes.GLASS]: textureGenerator.create({
      baseColor: "#c8dbe0",
      noise: { amount: 0.02, scale: 1 },
      border: true,
    }),

    [BlockTypes.BRICK]: textureGenerator.create({
      baseColor: "#b75a3c",
      noise: { amount: 0.1, scale: 1 },
      pattern: "brick",
      border: true,
    }),

    [BlockTypes.BEDROCK]: textureGenerator.create({
      baseColor: "#1a1a1a",
      noise: { amount: 0.2, scale: 1 },
      pattern: "stone",
    }),

    [BlockTypes.COBBLESTONE]: textureGenerator.create({
      baseColor: "#5a5a5a",
      noise: { amount: 0.2, scale: 1 },
      pattern: "stone",
      border: true,
    }),

    [BlockTypes.GRAVEL]: textureGenerator.create({
      baseColor: "#808080",
      noise: { amount: 0.25, scale: 1 },
      pattern: "stone",
    }),

    [BlockTypes.CLAY]: textureGenerator.create({
      baseColor: "#9fa4ad",
      noise: { amount: 0.1, scale: 1 },
      border: true,
    }),
  };
}

export default createBlockTextures;
