/**
 * Registro de blocos - Sistema extensível para adicionar novos tipos
 */

import { BlockTypes, BlockDefinitions } from "./BlockTypes.js";

class BlockRegistry {
  constructor() {
    this.blocks = new Map();
    this.nextId = 100; // IDs customizados começam em 100

    // Registra todos os blocos padrão
    Object.entries(BlockDefinitions).forEach(([id, def]) => {
      this.blocks.set(parseInt(id), {
        id: parseInt(id),
        ...def,
      });
    });
  }

  /**
   * Registra um novo tipo de bloco
   * @param {Object} definition - Definição do bloco
   * @returns {number} ID do bloco registrado
   */
  register(definition) {
    const id = definition.id || this.nextId++;

    this.blocks.set(id, {
      id,
      name: definition.name || "Unknown",
      solid: definition.solid !== false,
      transparent: definition.transparent || false,
      collidable: definition.collidable !== false,
      color: definition.color || "#ff00ff",
      opacity: definition.opacity || 1.0,
      emissive: definition.emissive || 0,
      breakTime: definition.breakTime || 1.0,
      textures: definition.textures || null,
      gravity: definition.gravity || false,
      liquid: definition.liquid || false,
      damages: definition.damages || false,
      indestructible: definition.indestructible || false,
      ...definition,
    });

    return id;
  }

  /**
   * Obtém definição de um bloco por ID
   */
  get(id) {
    return this.blocks.get(id) || this.blocks.get(BlockTypes.AIR);
  }

  /**
   * Verifica se bloco é sólido
   */
  isSolid(id) {
    const block = this.get(id);
    return block ? block.solid : false;
  }

  /**
   * Verifica se bloco é transparente
   */
  isTransparent(id) {
    const block = this.get(id);
    return block ? block.transparent : true;
  }

  /**
   * Verifica se bloco é colidível
   */
  isCollidable(id) {
    const block = this.get(id);
    return block ? block.collidable : false;
  }

  /**
   * Verifica se bloco é líquido
   */
  isLiquid(id) {
    const block = this.get(id);
    return block ? block.liquid : false;
  }

  /**
   * Obtém cor do bloco
   */
  getColor(id) {
    const block = this.get(id);
    return block ? block.color : "#ff00ff";
  }

  /**
   * Obtém nome do bloco
   */
  getName(id) {
    const block = this.get(id);
    return block ? block.name : "Unknown";
  }

  /**
   * Obtém opacidade do bloco
   */
  getOpacity(id) {
    const block = this.get(id);
    return block ? block.opacity : 1.0;
  }

  /**
   * Lista todos os blocos colocáveis pelo jogador
   */
  getPlaceableBlocks() {
    const placeable = [];
    this.blocks.forEach((block, id) => {
      if (id !== BlockTypes.AIR && id !== BlockTypes.BEDROCK && block.solid) {
        placeable.push(block);
      }
    });
    return placeable;
  }

  /**
   * Obtém todos os blocos de um tipo específico
   */
  getBlocksByProperty(property, value) {
    const result = [];
    this.blocks.forEach((block) => {
      if (block[property] === value) {
        result.push(block);
      }
    });
    return result;
  }
}

// Singleton
export const blockRegistry = new BlockRegistry();
export default blockRegistry;
