import path from "node:path"
import react from "@vitejs/plugin-react"
import tailwindcss from "@tailwindcss/vite"
import { defineConfig } from "vite"

// In dev, API calls are proxied to a running daemon (`nbx daemon start`).
// Sign in once by opening the URL from `nbx open -print-url` on the daemon's
// port; the dev server forwards the cookie.
const daemon = process.env.NBX_DAEMON ?? "http://127.0.0.1:8765"

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: { "@": path.resolve(__dirname, "./src") },
  },
  build: {
    outDir: "../internal/web/dist",
    emptyOutDir: true,
    chunkSizeWarningLimit: 4000,
  },
  server: {
    proxy: {
      "/api": {
        target: daemon,
        changeOrigin: true,
      },
    },
  },
})
