/// <reference types="vitest" />

import path from "node:path";

import react from "@vitejs/plugin-react";
import { defineConfig, type PluginOption } from "vite";
import { visualizer } from "rollup-plugin-visualizer";

const analyze = process.env.ANALYZE === "true";

const plugins: PluginOption[] = [react()];

if (analyze) {
  plugins.push(
    visualizer({
      filename: "dist/stats.html",
      gzipSize: true,
      brotliSize: true,
      template: "treemap",
      open: true
    }) as PluginOption
  );
}

export default defineConfig({
  plugins,
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src")
    }
  },
  server: {
    port: 3000,
    host: "127.0.0.1",
    proxy: {
      "/api": {
        target: process.env.SING_GROK_API_BASE_URL || "http://127.0.0.1:8081",
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: "dist",
    sourcemap: analyze
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: "./src/test/setup.ts",
    css: true,
    exclude: ["node_modules/**", "dist/**", "e2e/**"]
  }
});
