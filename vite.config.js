import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  base: "./", // Important for Electron production build
  resolve: {
    alias: {
      "@": "/src",
      "@core": "/src/core",
      "@generation": "/src/generation",
      "@rendering": "/src/rendering",
      "@controls": "/src/controls",
      "@entities": "/src/entities",
      "@ui": "/src/ui",
      "@utils": "/src/utils",
    },
  },
  optimizeDeps: {
    include: ["three", "@react-three/fiber", "@react-three/drei"],
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
  server: {
    port: 5173,
    strictPort: false,
  },
});
