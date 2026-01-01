/**
 * SaveManager - Handles game save/load functionality
 * Saves: player position, inventory, world modifications (chunks)
 */

// Check if running in Electron
const isElectron =
  typeof window !== "undefined" && window.electronAPI?.isElectron;

/**
 * Serializes chunk modifications for saving
 * Only saves blocks that differ from procedurally generated terrain
 */
function serializeChunkModifications(chunkManager, terrainGenerator) {
  const modifiedChunks = {};

  for (const [id, chunk] of chunkManager.chunks) {
    const modifications = [];

    // Create a temporary chunk to compare against original generation
    // This is expensive but ensures we only save actual modifications
    for (let x = 0; x < 16; x++) {
      for (let z = 0; z < 16; z++) {
        for (let y = 0; y < 64; y++) {
          const currentBlock = chunk.getBlock(x, y, z);
          // Track all non-air blocks for simplicity (can optimize later)
          if (chunk.modified && chunk.modifiedBlocks) {
            // If chunk tracks modifications, use that
            const key = `${x},${y},${z}`;
            if (chunk.modifiedBlocks.has(key)) {
              modifications.push({ x, y, z, type: currentBlock });
            }
          }
        }
      }
    }

    if (modifications.length > 0) {
      modifiedChunks[id] = {
        cx: chunk.cx,
        cz: chunk.cz,
        modifications,
      };
    }
  }

  // Also check cached chunks
  if (chunkManager.chunkCache) {
    for (const [id, chunk] of chunkManager.chunkCache.entries?.() || []) {
      if (chunk.modifiedBlocks && chunk.modifiedBlocks.size > 0) {
        const modifications = [];
        for (const [key, type] of chunk.modifiedBlocks) {
          const [x, y, z] = key.split(",").map(Number);
          modifications.push({ x, y, z, type });
        }
        modifiedChunks[id] = {
          cx: chunk.cx,
          cz: chunk.cz,
          modifications,
        };
      }
    }
  }

  return modifiedChunks;
}

/**
 * Creates a save game object
 */
export function createSaveData(
  world,
  playerPosition,
  playerRotation,
  inventory
) {
  return {
    version: "1.0",
    timestamp: Date.now(),
    player: {
      position: {
        x: playerPosition.x,
        y: playerPosition.y,
        z: playerPosition.z,
      },
      rotation: playerRotation || 0,
    },
    inventory: inventory
      ? {
          slots: inventory.getAll ? inventory.getAll() : [],
        }
      : { slots: [] },
    world: {
      seed: world.seed,
      modifiedChunks: serializeChunkModifications(
        world.chunkManager,
        world.terrainGenerator
      ),
    },
  };
}

/**
 * Saves game to local storage (fallback for browser)
 */
export function saveToLocalStorage(saveData, saveName = "quicksave") {
  try {
    const key = `voxel-save-${saveName}`;
    localStorage.setItem(key, JSON.stringify(saveData));
    console.log(`[SaveManager] Saved to localStorage: ${key}`);
    return true;
  } catch (error) {
    console.error("[SaveManager] Failed to save to localStorage:", error);
    return false;
  }
}

/**
 * Loads game from local storage (fallback for browser)
 */
export function loadFromLocalStorage(saveName = "quicksave") {
  try {
    const key = `voxel-save-${saveName}`;
    const data = localStorage.getItem(key);
    if (!data) {
      console.log(`[SaveManager] No save found: ${key}`);
      return null;
    }
    const saveData = JSON.parse(data);
    console.log(`[SaveManager] Loaded from localStorage: ${key}`);
    return saveData;
  } catch (error) {
    console.error("[SaveManager] Failed to load from localStorage:", error);
    return null;
  }
}

/**
 * Saves game using Electron IPC (preferred)
 */
export async function saveToFile(saveData, saveName = "quicksave") {
  if (!isElectron) {
    return saveToLocalStorage(saveData, saveName);
  }

  try {
    const result = await window.electronAPI.saveGame(saveName, saveData);
    console.log(`[SaveManager] Saved to file: ${saveName}`);
    return result;
  } catch (error) {
    console.error("[SaveManager] Failed to save to file:", error);
    // Fallback to localStorage
    return saveToLocalStorage(saveData, saveName);
  }
}

/**
 * Loads game using Electron IPC (preferred)
 */
export async function loadFromFile(saveName = "quicksave") {
  if (!isElectron) {
    return loadFromLocalStorage(saveName);
  }

  try {
    const saveData = await window.electronAPI.loadGame(saveName);
    console.log(`[SaveManager] Loaded from file: ${saveName}`);
    return saveData;
  } catch (error) {
    console.error("[SaveManager] Failed to load from file:", error);
    // Fallback to localStorage
    return loadFromLocalStorage(saveName);
  }
}

/**
 * Applies loaded save data to the world
 */
export function applySaveData(saveData, world, setPlayerPosition, inventory) {
  if (!saveData) return false;

  try {
    // Restore player position
    if (saveData.player?.position) {
      setPlayerPosition({
        x: saveData.player.position.x,
        y: saveData.player.position.y,
        z: saveData.player.position.z,
      });
    }

    // Restore inventory
    if (saveData.inventory?.slots && inventory) {
      saveData.inventory.slots.forEach((item, index) => {
        if (item) {
          inventory.setSlot(index, item);
        }
      });
    }

    // Restore chunk modifications
    if (saveData.world?.modifiedChunks) {
      for (const [id, chunkData] of Object.entries(
        saveData.world.modifiedChunks
      )) {
        // Get or create the chunk
        let chunk = world.chunkManager.getChunk(chunkData.cx, chunkData.cz);

        if (!chunk) {
          // Load chunk from cache or generate
          chunk = world.chunkManager.loadChunk(chunkData.cx, chunkData.cz);
        }

        if (chunk && chunkData.modifications) {
          // Apply modifications
          chunkData.modifications.forEach((mod) => {
            chunk.setBlock(mod.x, mod.y, mod.z, mod.type);
            // Track as modified
            if (!chunk.modifiedBlocks) {
              chunk.modifiedBlocks = new Map();
            }
            chunk.modifiedBlocks.set(`${mod.x},${mod.y},${mod.z}`, mod.type);
          });
          chunk.isDirty = true;
        }
      }
    }

    console.log("[SaveManager] Save data applied successfully");
    return true;
  } catch (error) {
    console.error("[SaveManager] Failed to apply save data:", error);
    return false;
  }
}

/**
 * Lists available saves
 */
export async function listSaves() {
  if (!isElectron) {
    // Browser fallback - check localStorage
    const saves = [];
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key?.startsWith("voxel-save-")) {
        const saveName = key.replace("voxel-save-", "");
        try {
          const data = JSON.parse(localStorage.getItem(key));
          saves.push({
            name: saveName,
            timestamp: data.timestamp,
            version: data.version,
          });
        } catch {}
      }
    }
    return saves;
  }

  try {
    return await window.electronAPI.listSaves();
  } catch (error) {
    console.error("[SaveManager] Failed to list saves:", error);
    return [];
  }
}

export default {
  createSaveData,
  saveToFile,
  loadFromFile,
  applySaveData,
  listSaves,
};
