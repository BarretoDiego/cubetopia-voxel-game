/**
 * Gerador de Criaturas Procedurais
 * Cria entidades com partes corporais modulares
 */

import { SeededRNG } from "../../utils/mathUtils.js";

// Templates de criaturas
export const CreatureTemplates = {
  QUADRUPED: "quadruped", // 4 pernas (animais terrestres)
  BIPED: "biped", // 2 pernas (humanoides)
  FLYING: "flying", // Asas (pássaros, morcegos)
  SLIME: "slime", // Blob (slimes, gelatinas)
  FISH: "fish", // Aquático
  SPIDER: "spider", // 8 pernas
};

// Cores de criaturas por bioma
export const BiomeCreatureColors = {
  plains: ["#8B4513", "#DAA520", "#F5DEB3", "#D2B48C", "#A0522D"],
  forest: ["#228B22", "#8B4513", "#2E8B57", "#6B8E23", "#556B2F"],
  desert: ["#E0C090", "#DEB887", "#F4A460", "#D2691E", "#CD853F"],
  snow: ["#FFFFFF", "#F0F8FF", "#E6E6FA", "#B0C4DE", "#778899"],
  mountains: ["#696969", "#808080", "#A9A9A9", "#2F4F4F", "#708090"],
};

// Comportamentos disponíveis
export const Behaviors = {
  WANDER: "wander",
  FOLLOW: "follow",
  FLEE: "flee",
  IDLE: "idle",
  JUMP: "jump",
  SWIM: "swim",
};

/**
 * Classe para criar criaturas procedurais
 */
export class CreatureGenerator {
  constructor(seed = Date.now()) {
    this.seed = seed;
    this.rng = new SeededRNG(seed);
    this.creatureId = 0;
  }

  /**
   * Cria uma criatura procedural
   */
  create(options = {}) {
    const template =
      options.template || this.rng.choose(Object.values(CreatureTemplates));
    const biome = options.biome || "plains";
    const size = options.size || this.rng.nextFloat(0.5, 2.0);

    const creature = {
      id: this.creatureId++,
      template,
      size,
      biome,
      // Propriedades físicas
      body: this._generateBody(template, size),
      // Aparência
      colors: this._generateColors(biome),
      // Comportamento
      behaviors: this._generateBehaviors(template),
      // Estatísticas
      stats: this._generateStats(template, size),
      // Posição e movimento
      position: options.position || { x: 0, y: 0, z: 0 },
      velocity: { x: 0, y: 0, z: 0 },
      rotation: 0,
      // Estado
      state: "idle",
      target: null,
      timer: 0,
    };

    return creature;
  }

  /**
   * Gera corpo da criatura
   */
  _generateBody(template, size) {
    const body = {
      parts: [],
      height: size,
      width: size * 0.6,
      depth: size * 0.8,
    };

    switch (template) {
      case CreatureTemplates.QUADRUPED:
        body.parts = [
          {
            type: "torso",
            size: { x: 0.8, y: 0.5, z: 1.2 },
            offset: { x: 0, y: 0.5, z: 0 },
          },
          {
            type: "head",
            size: { x: 0.4, y: 0.4, z: 0.5 },
            offset: { x: 0, y: 0.7, z: 0.6 },
          },
          {
            type: "leg",
            size: { x: 0.15, y: 0.5, z: 0.15 },
            offset: { x: -0.3, y: 0, z: 0.4 },
          },
          {
            type: "leg",
            size: { x: 0.15, y: 0.5, z: 0.15 },
            offset: { x: 0.3, y: 0, z: 0.4 },
          },
          {
            type: "leg",
            size: { x: 0.15, y: 0.5, z: 0.15 },
            offset: { x: -0.3, y: 0, z: -0.4 },
          },
          {
            type: "leg",
            size: { x: 0.15, y: 0.5, z: 0.15 },
            offset: { x: 0.3, y: 0, z: -0.4 },
          },
          {
            type: "tail",
            size: { x: 0.1, y: 0.1, z: 0.4 },
            offset: { x: 0, y: 0.5, z: -0.8 },
          },
        ];
        break;

      case CreatureTemplates.BIPED:
        body.parts = [
          {
            type: "torso",
            size: { x: 0.5, y: 0.7, z: 0.3 },
            offset: { x: 0, y: 0.8, z: 0 },
          },
          {
            type: "head",
            size: { x: 0.35, y: 0.35, z: 0.35 },
            offset: { x: 0, y: 1.35, z: 0 },
          },
          {
            type: "arm",
            size: { x: 0.15, y: 0.6, z: 0.15 },
            offset: { x: -0.35, y: 0.9, z: 0 },
          },
          {
            type: "arm",
            size: { x: 0.15, y: 0.6, z: 0.15 },
            offset: { x: 0.35, y: 0.9, z: 0 },
          },
          {
            type: "leg",
            size: { x: 0.2, y: 0.7, z: 0.2 },
            offset: { x: -0.15, y: 0, z: 0 },
          },
          {
            type: "leg",
            size: { x: 0.2, y: 0.7, z: 0.2 },
            offset: { x: 0.15, y: 0, z: 0 },
          },
        ];
        break;

      case CreatureTemplates.FLYING:
        body.parts = [
          {
            type: "body",
            size: { x: 0.3, y: 0.3, z: 0.5 },
            offset: { x: 0, y: 0, z: 0 },
          },
          {
            type: "head",
            size: { x: 0.2, y: 0.2, z: 0.25 },
            offset: { x: 0, y: 0.1, z: 0.3 },
          },
          {
            type: "wing",
            size: { x: 0.8, y: 0.05, z: 0.3 },
            offset: { x: -0.5, y: 0.1, z: 0 },
          },
          {
            type: "wing",
            size: { x: 0.8, y: 0.05, z: 0.3 },
            offset: { x: 0.5, y: 0.1, z: 0 },
          },
          {
            type: "tail",
            size: { x: 0.15, y: 0.05, z: 0.4 },
            offset: { x: 0, y: 0, z: -0.4 },
          },
        ];
        break;

      case CreatureTemplates.SLIME:
        const blobSize = 0.5 + this.rng.nextFloat(0, 0.5);
        body.parts = [
          {
            type: "blob",
            size: { x: blobSize, y: blobSize * 0.8, z: blobSize },
            offset: { x: 0, y: blobSize * 0.4, z: 0 },
          },
        ];
        break;

      case CreatureTemplates.FISH:
        body.parts = [
          {
            type: "body",
            size: { x: 0.2, y: 0.3, z: 0.6 },
            offset: { x: 0, y: 0, z: 0 },
          },
          {
            type: "fin",
            size: { x: 0.3, y: 0.2, z: 0.1 },
            offset: { x: 0, y: 0.2, z: 0 },
          },
          {
            type: "tail",
            size: { x: 0.05, y: 0.25, z: 0.2 },
            offset: { x: 0, y: 0, z: -0.35 },
          },
        ];
        break;

      case CreatureTemplates.SPIDER:
        body.parts = [
          {
            type: "abdomen",
            size: { x: 0.5, y: 0.4, z: 0.6 },
            offset: { x: 0, y: 0.3, z: -0.3 },
          },
          {
            type: "thorax",
            size: { x: 0.3, y: 0.25, z: 0.3 },
            offset: { x: 0, y: 0.25, z: 0.2 },
          },
          {
            type: "head",
            size: { x: 0.2, y: 0.2, z: 0.2 },
            offset: { x: 0, y: 0.25, z: 0.4 },
          },
        ];
        // Adiciona 8 pernas
        for (let i = 0; i < 8; i++) {
          const side = i % 2 === 0 ? -1 : 1;
          const zOffset = (Math.floor(i / 2) - 1.5) * 0.15;
          body.parts.push({
            type: "leg",
            size: { x: 0.3, y: 0.05, z: 0.05 },
            offset: { x: side * 0.3, y: 0.2, z: zOffset },
          });
        }
        break;
    }

    // Escala todas as partes pelo tamanho
    body.parts = body.parts.map((part) => ({
      ...part,
      size: {
        x: part.size.x * size,
        y: part.size.y * size,
        z: part.size.z * size,
      },
      offset: {
        x: part.offset.x * size,
        y: part.offset.y * size,
        z: part.offset.z * size,
      },
    }));

    return body;
  }

  /**
   * Gera cores para a criatura
   */
  _generateColors(biome) {
    const palette = BiomeCreatureColors[biome] || BiomeCreatureColors.plains;
    const primary = this.rng.choose(palette);
    const secondary = this.rng.choose(palette);

    return {
      primary,
      secondary,
      accent: this._lightenColor(primary, 0.3),
      eyes: "#000000",
    };
  }

  /**
   * Gera comportamentos para a criatura
   */
  _generateBehaviors(template) {
    const behaviors = [Behaviors.IDLE, Behaviors.WANDER];

    switch (template) {
      case CreatureTemplates.QUADRUPED:
        behaviors.push(Behaviors.FLEE);
        break;
      case CreatureTemplates.BIPED:
        behaviors.push(Behaviors.FOLLOW);
        break;
      case CreatureTemplates.FLYING:
        behaviors.push(Behaviors.FLEE);
        break;
      case CreatureTemplates.SLIME:
        behaviors.push(Behaviors.JUMP);
        break;
      case CreatureTemplates.FISH:
        behaviors.push(Behaviors.SWIM);
        break;
      case CreatureTemplates.SPIDER:
        behaviors.push(Behaviors.FOLLOW);
        break;
    }

    return behaviors;
  }

  /**
   * Gera estatísticas da criatura
   */
  _generateStats(template, size) {
    const baseHealth = 10;
    const baseSpeed = 2;

    return {
      health: Math.floor(baseHealth * size),
      maxHealth: Math.floor(baseHealth * size),
      speed: baseSpeed / size, // Criaturas menores são mais rápidas
      jumpForce: template === CreatureTemplates.SLIME ? 8 : 5,
      damage: Math.floor(2 * size),
      hostile: template === CreatureTemplates.SPIDER,
    };
  }

  /**
   * Clareia uma cor hex
   */
  _lightenColor(hex, amount) {
    const num = parseInt(hex.slice(1), 16);
    const r = Math.min(255, ((num >> 16) & 0xff) + 255 * amount);
    const g = Math.min(255, ((num >> 8) & 0xff) + 255 * amount);
    const b = Math.min(255, (num & 0xff) + 255 * amount);
    return `#${(
      (1 << 24) +
      (Math.floor(r) << 16) +
      (Math.floor(g) << 8) +
      Math.floor(b)
    )
      .toString(16)
      .slice(1)}`;
  }

  /**
   * Atualiza o comportamento de uma criatura
   */
  updateBehavior(creature, deltaTime, world, playerPosition) {
    creature.timer += deltaTime;

    switch (creature.state) {
      case "idle":
        if (creature.timer > 2 + Math.random() * 3) {
          creature.state = "wander";
          creature.timer = 0;
          creature.target = {
            x: creature.position.x + (Math.random() - 0.5) * 10,
            z: creature.position.z + (Math.random() - 0.5) * 10,
          };
        }
        break;

      case "wander":
        if (creature.target) {
          const dx = creature.target.x - creature.position.x;
          const dz = creature.target.z - creature.position.z;
          const dist = Math.sqrt(dx * dx + dz * dz);

          if (dist < 0.5 || creature.timer > 5) {
            creature.state = "idle";
            creature.timer = 0;
            creature.target = null;
          } else {
            creature.velocity.x = (dx / dist) * creature.stats.speed;
            creature.velocity.z = (dz / dist) * creature.stats.speed;
            creature.rotation = Math.atan2(dx, dz);
          }
        }
        break;

      case "flee":
        if (playerPosition) {
          const dx = creature.position.x - playerPosition.x;
          const dz = creature.position.z - playerPosition.z;
          const dist = Math.sqrt(dx * dx + dz * dz);

          if (dist > 15) {
            creature.state = "idle";
            creature.timer = 0;
          } else {
            creature.velocity.x = (dx / dist) * creature.stats.speed * 1.5;
            creature.velocity.z = (dz / dist) * creature.stats.speed * 1.5;
            creature.rotation = Math.atan2(-dx, -dz);
          }
        }
        break;
    }

    // Slimes pulam periodicamente
    if (creature.template === CreatureTemplates.SLIME && creature.timer > 1) {
      if (creature.position.y < 0.1) {
        creature.velocity.y = creature.stats.jumpForce;
        creature.timer = 0;
      }
    }

    return creature;
  }
}

export default CreatureGenerator;
