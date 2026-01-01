/**
 * Water Simulator - Dynamic water flow physics
 * Water flows down and spreads horizontally
 */

import { BlockTypes } from "../core/blocks/BlockTypes.js";

export class WaterSimulator {
  constructor(world) {
    this.world = world;
    this.updateQueue = new Set();
    this.maxSpread = 7; // How far water spreads horizontally
  }

  /**
   * Queue a block position for water update
   */
  queueUpdate(x, y, z) {
    this.updateQueue.add(`${x},${y},${z}`);
  }

  /**
   * Process water updates (call this from game loop)
   * @param maxUpdates - Maximum updates per tick to prevent lag
   */
  tick(maxUpdates = 10) {
    let processed = 0;
    const toProcess = Array.from(this.updateQueue);

    for (const key of toProcess) {
      if (processed >= maxUpdates) break;

      const [x, y, z] = key.split(",").map(Number);
      this.updateQueue.delete(key);

      if (this.processWater(x, y, z)) {
        processed++;
      }
    }
  }

  /**
   * Process water at a specific position
   */
  processWater(x, y, z) {
    const block = this.world.getBlock(x, y, z);

    // Only process water blocks
    if (block !== BlockTypes.WATER) return false;

    let updated = false;

    // Flow down first (priority)
    const below = this.world.getBlock(x, y - 1, z);
    if (below === BlockTypes.AIR) {
      this.world.setBlock(x, y - 1, z, BlockTypes.WATER);
      this.queueUpdate(x, y - 1, z);
      updated = true;
    }

    // If can't flow down, spread horizontally
    if (!updated || below === BlockTypes.WATER) {
      const directions = [
        [1, 0],
        [-1, 0],
        [0, 1],
        [0, -1],
      ];

      for (const [dx, dz] of directions) {
        const nx = x + dx;
        const nz = z + dz;
        const neighbor = this.world.getBlock(nx, y, nz);

        if (neighbor === BlockTypes.AIR) {
          // Check if there's support below or water below
          const belowNeighbor = this.world.getBlock(nx, y - 1, nz);
          if (belowNeighbor !== BlockTypes.AIR) {
            this.world.setBlock(nx, y, nz, BlockTypes.WATER);
            this.queueUpdate(nx, y, nz);
            updated = true;
          }
        }
      }
    }

    return updated;
  }

  /**
   * Trigger update when a block is removed
   */
  onBlockRemoved(x, y, z) {
    // Check if any neighboring water should flow into this space
    const neighbors = [
      [x, y + 1, z], // Above
      [x + 1, y, z], // East
      [x - 1, y, z], // West
      [x, y, z + 1], // South
      [x, y, z - 1], // North
    ];

    for (const [nx, ny, nz] of neighbors) {
      if (this.world.getBlock(nx, ny, nz) === BlockTypes.WATER) {
        this.queueUpdate(nx, ny, nz);
      }
    }
  }

  /**
   * Called when water block is placed
   */
  onWaterPlaced(x, y, z) {
    this.queueUpdate(x, y, z);
  }
}

export default WaterSimulator;
