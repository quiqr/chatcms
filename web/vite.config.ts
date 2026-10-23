import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// scripts/dev.sh serves on 8080 unless CHATCMS_PORT says otherwise.
const apiTarget =
  process.env.CHATCMS_API_TARGET ?? `http://127.0.0.1:${process.env.CHATCMS_PORT ?? "8080"}`;

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      "/api": {
        target: apiTarget,
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./vitest.setup.ts"],
    css: false,
  },
});
