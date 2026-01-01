// Script to wait for Vite dev server before starting Electron
// Tries multiple ports since Vite may fall back to another port
import http from "http";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const MAX_ATTEMPTS = 30;
const INTERVAL = 1000;
const PORTS_TO_TRY = [5173, 5174, 5175, 5176, 5177];

async function checkPort(port) {
  return new Promise((resolve) => {
    const req = http.get(`http://localhost:${port}`, (res) => {
      resolve(res.statusCode === 200);
    });
    req.on("error", () => resolve(false));
    req.setTimeout(500, () => {
      req.destroy();
      resolve(false);
    });
  });
}

async function waitForVite() {
  for (let i = 0; i < MAX_ATTEMPTS; i++) {
    for (const port of PORTS_TO_TRY) {
      const isReady = await checkPort(port);
      if (isReady) {
        console.log(`Vite server is ready on port ${port}!`);

        // Write port to temp file for Electron main process to read
        const portFile = path.join(__dirname, "..", ".vite-port");
        fs.writeFileSync(portFile, String(port));

        return;
      }
    }
    console.log(`Waiting for Vite server... (${i + 1}/${MAX_ATTEMPTS})`);
    await new Promise((r) => setTimeout(r, INTERVAL));
  }
  console.error("Vite server did not start in time");
  process.exit(1);
}

waitForVite();
