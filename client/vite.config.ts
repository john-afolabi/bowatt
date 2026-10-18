import path from "node:path";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite-plus";

const API_TARGET = process.env.API_PROXY_TARGET ?? "http://localhost:8080";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "./src"),
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/api": { target: API_TARGET, changeOrigin: false, ws: true },
      "/uploads": { target: API_TARGET, changeOrigin: false },
      "/healthz": { target: API_TARGET, changeOrigin: false },
    },
  },
  test: {
    include: ["src/**/*.test.{ts,tsx}"],
    environment: "happy-dom",
    restoreMocks: true,
  },
  lint: {
    ignorePatterns: ["dist/**", "node_modules/**", "src/components/ui/**"],
    options: {
      // Turning this off would silently stop `vp check` type checking anything.
      typeAware: true,
      typeCheck: true,
    },
  },
  fmt: {
    ignorePatterns: ["dist/**", "node_modules/**", "src/components/ui/**"],
  },
});
