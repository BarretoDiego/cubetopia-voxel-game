const { app, BrowserWindow, dialog, ipcMain } = require("electron");
const path = require("path");
const fs = require("fs");

// ============================================================================
// SAVE SYSTEM - Directory for game saves
// ============================================================================

const SAVES_DIR = path.join(app.getPath("userData"), "saves");

// Ensure saves directory exists
function ensureSavesDir() {
  if (!fs.existsSync(SAVES_DIR)) {
    fs.mkdirSync(SAVES_DIR, { recursive: true });
  }
}

// IPC Handlers for Save/Load
ipcMain.handle("save-game", async (event, saveName, saveData) => {
  try {
    ensureSavesDir();
    const savePath = path.join(SAVES_DIR, `${saveName}.json`);
    fs.writeFileSync(savePath, JSON.stringify(saveData, null, 2));
    console.log(`[SaveSystem] Saved game to: ${savePath}`);
    return { success: true, path: savePath };
  } catch (error) {
    console.error("[SaveSystem] Failed to save:", error);
    return { success: false, error: error.message };
  }
});

ipcMain.handle("load-game", async (event, saveName) => {
  try {
    ensureSavesDir();
    const savePath = path.join(SAVES_DIR, `${saveName}.json`);
    if (!fs.existsSync(savePath)) {
      return null;
    }
    const data = fs.readFileSync(savePath, "utf-8");
    console.log(`[SaveSystem] Loaded game from: ${savePath}`);
    return JSON.parse(data);
  } catch (error) {
    console.error("[SaveSystem] Failed to load:", error);
    return null;
  }
});

ipcMain.handle("list-saves", async () => {
  try {
    ensureSavesDir();
    const files = fs.readdirSync(SAVES_DIR).filter((f) => f.endsWith(".json"));
    const saves = files.map((file) => {
      try {
        const data = JSON.parse(
          fs.readFileSync(path.join(SAVES_DIR, file), "utf-8")
        );
        return {
          name: file.replace(".json", ""),
          timestamp: data.timestamp,
          version: data.version,
        };
      } catch {
        return { name: file.replace(".json", ""), timestamp: 0 };
      }
    });
    return saves;
  } catch (error) {
    console.error("[SaveSystem] Failed to list saves:", error);
    return [];
  }
});

ipcMain.handle("delete-save", async (event, saveName) => {
  try {
    const savePath = path.join(SAVES_DIR, `${saveName}.json`);
    if (fs.existsSync(savePath)) {
      fs.unlinkSync(savePath);
      return true;
    }
    return false;
  } catch (error) {
    console.error("[SaveSystem] Failed to delete save:", error);
    return false;
  }
});

// ============================================================================
// CHUNK CACHE - Disk-based chunk caching to reduce memory
// ============================================================================

const CHUNKS_DIR = path.join(app.getPath("userData"), "chunks");

function ensureChunksDir() {
  if (!fs.existsSync(CHUNKS_DIR)) {
    fs.mkdirSync(CHUNKS_DIR, { recursive: true });
  }
}

// Save chunk to disk (for caching)
ipcMain.handle("cache-chunk", async (event, chunkId, chunkData) => {
  try {
    ensureChunksDir();
    const chunkPath = path.join(
      CHUNKS_DIR,
      `${chunkId.replace(",", "_")}.json`
    );
    fs.writeFileSync(chunkPath, JSON.stringify(chunkData));
    return true;
  } catch (error) {
    console.error("[ChunkCache] Failed to cache chunk:", error);
    return false;
  }
});

// Load chunk from disk cache
ipcMain.handle("load-cached-chunk", async (event, chunkId) => {
  try {
    ensureChunksDir();
    const chunkPath = path.join(
      CHUNKS_DIR,
      `${chunkId.replace(",", "_")}.json`
    );
    if (!fs.existsSync(chunkPath)) {
      return null;
    }
    const data = fs.readFileSync(chunkPath, "utf-8");
    return JSON.parse(data);
  } catch (error) {
    console.error("[ChunkCache] Failed to load cached chunk:", error);
    return null;
  }
});

// Clear chunk cache (for new game)
ipcMain.handle("clear-chunk-cache", async () => {
  try {
    ensureChunksDir();
    const files = fs.readdirSync(CHUNKS_DIR);
    for (const file of files) {
      fs.unlinkSync(path.join(CHUNKS_DIR, file));
    }
    return true;
  } catch (error) {
    console.error("[ChunkCache] Failed to clear cache:", error);
    return false;
  }
});

// ============================================================================
// MEMORY LIMITS - Prevent system freeze
// ============================================================================

// Limit V8 heap memory to 2GB
app.commandLine.appendSwitch("js-flags", "--max-old-space-size=2048");

// Reduce Chromium's memory usage
app.commandLine.appendSwitch("disable-software-rasterizer");

// Disable hardware acceleration to reduce GPU memory (enable if still having issues)
// app.disableHardwareAcceleration();

const isDev = !app.isPackaged;

let mainWindow = null;

function getVitePort() {
  try {
    const portFile = path.join(__dirname, "..", ".vite-port");
    return parseInt(fs.readFileSync(portFile, "utf-8").trim(), 10);
  } catch {
    return 5173; // fallback
  }
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 720,
    minWidth: 800,
    minHeight: 600,
    title: "Voxel World Engine",
    backgroundColor: "#000000",
    show: false,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: true,
      preload: path.join(__dirname, "preload.cjs"),
      backgroundThrottling: false, // Keep game running in background
      devTools: isDev,
    },
  });

  // Show window when ready
  mainWindow.once("ready-to-show", () => {
    mainWindow.show();
    if (isDev) {
      mainWindow.webContents.openDevTools();
    }
  });

  // Load the app
  if (isDev) {
    // Development: load from Vite dev server
    const port = getVitePort();
    mainWindow.loadURL(`http://localhost:${port}`);
  } else {
    // Production: load built files
    mainWindow.loadFile(path.join(__dirname, "../dist/index.html"));
  }

  // Handle renderer process crash
  mainWindow.webContents.on("render-process-gone", (event, details) => {
    console.error("Renderer process crashed:", details);

    if (details.reason === "oom") {
      dialog.showErrorBox(
        "Memória Insuficiente",
        "O jogo excedeu o limite de memória de 4GB e foi encerrado.\n" +
          "Tente reduzir a distância de renderização ou reiniciar o jogo."
      );
    } else {
      dialog.showErrorBox(
        "Erro no Jogo",
        `O processo de renderização falhou: ${details.reason}\n` +
          "O aplicativo será reiniciado."
      );
    }

    // Restart the app
    app.relaunch();
    app.exit(0);
  });

  // Handle unresponsive
  mainWindow.webContents.on("unresponsive", () => {
    const response = dialog.showMessageBoxSync(mainWindow, {
      type: "warning",
      title: "Aplicativo Não Respondendo",
      message: "O jogo não está respondendo. Deseja aguardar ou reiniciar?",
      buttons: ["Aguardar", "Reiniciar"],
      defaultId: 0,
    });

    if (response === 1) {
      mainWindow.destroy();
      createWindow();
    }
  });

  // Cleanup on close
  mainWindow.on("closed", () => {
    mainWindow = null;
  });
}

// App lifecycle
app.whenReady().then(() => {
  createWindow();

  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});

// Prevent multiple instances
const gotTheLock = app.requestSingleInstanceLock();
if (!gotTheLock) {
  app.quit();
} else {
  app.on("second-instance", () => {
    if (mainWindow) {
      if (mainWindow.isMinimized()) mainWindow.restore();
      mainWindow.focus();
    }
  });
}
