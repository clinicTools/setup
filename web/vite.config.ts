import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";

// Das Frontend wird direkt in das Go-Binary eingebettet, daher landet der
// Build-Output im embed-Verzeichnis internal/web/dist.
export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  build: {
    outDir: "../internal/web/dist",
    emptyOutDir: true,
  },
  server: {
    port: 5173,
    proxy: {
      // Im Dev-Modus an den Go-Backend-Dienst weiterleiten.
      "/api": {
        target: "http://localhost:8088",
        changeOrigin: true,
        ws: true, // WebSocket-Upgrade (/api/ws) durchreichen
      },
    },
  },
});
