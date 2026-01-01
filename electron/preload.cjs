const { contextBridge, ipcRenderer } = require("electron");

// Expose protected methods for game functionality
contextBridge.exposeInMainWorld("electronAPI", {
  // Check if running in Electron
  isElectron: true,

  // Platform info
  platform: process.platform,

  // Get current memory usage
  getMemoryUsage: () => {
    return process.memoryUsage();
  },

  // Get memory limit (4GB in bytes)
  getMemoryLimit: () => {
    return 4 * 1024 * 1024 * 1024; // 4GB
  },

  // ============================================
  // SAVE SYSTEM APIs
  // ============================================

  // Save game to file
  saveGame: (saveName, saveData) => {
    return ipcRenderer.invoke("save-game", saveName, saveData);
  },

  // Load game from file
  loadGame: (saveName) => {
    return ipcRenderer.invoke("load-game", saveName);
  },

  // List all saves
  listSaves: () => {
    return ipcRenderer.invoke("list-saves");
  },

  // Delete a save
  deleteSave: (saveName) => {
    return ipcRenderer.invoke("delete-save", saveName);
  },

  // ============================================
  // CHUNK CACHE APIs (Disk-based caching)
  // ============================================

  // Cache chunk to disk
  cacheChunk: (chunkId, chunkData) => {
    return ipcRenderer.invoke("cache-chunk", chunkId, chunkData);
  },

  // Load cached chunk from disk
  loadCachedChunk: (chunkId) => {
    return ipcRenderer.invoke("load-cached-chunk", chunkId);
  },

  // Clear all cached chunks
  clearChunkCache: () => {
    return ipcRenderer.invoke("clear-chunk-cache");
  },
});
