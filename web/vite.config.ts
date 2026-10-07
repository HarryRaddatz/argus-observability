/// <reference types="vitest/config" />
import path from "path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "./src"),
    },
  },
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
  },
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: process.env.VITE_HUB_PROXY ?? "http://127.0.0.1:8080",
        changeOrigin: true,
        configure: (proxy) => {
          proxy.on("proxyReq", (proxyReq) => {
            const token = process.env.ARGUS_AGENT_TOKEN
            if (token) proxyReq.setHeader("Authorization", `Bearer ${token}`)
          })
        },
      },
      "/health": {
        target: process.env.VITE_HUB_PROXY ?? "http://127.0.0.1:8080",
        changeOrigin: true,
      },
    },
  },
})
