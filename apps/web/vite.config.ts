import { fileURLToPath, URL } from "node:url"
import vue from "@vitejs/plugin-vue"
import { defineConfig } from "vite"

// In development, /api is proxied to the Go backend (same origin for the browser).
export default defineConfig({
  plugins: [vue()],
  resolve: { alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) } },
  server: {
    port: 5173,
    proxy: { "/api": { target: process.env.BACKEND_URL ?? "http://localhost:5100", changeOrigin: true } },
  },
  test: { environment: "jsdom" },
})
