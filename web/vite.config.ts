import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// scripts/dev.sh serves on 8080 unless CHATCMS_PORT says otherwise.
const apiTarget =
  process.env.CHATCMS_API_TARGET ?? `http://127.0.0.1:${process.env.CHATCMS_PORT ?? "8080"}`;

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    // The same setting the server reads, so that one variable makes both
    // halves reachable from another machine rather than only the half that
    // happens to be listening.
    host: process.env.CHATCMS_HOST || undefined,
    proxy: {
      "/api": {
        target: apiTarget,
        changeOrigin: true,
      },
      // The preview is served by the Go server under the app's own address,
      // so the development server has to forward it too, websocket and all:
      // Hugo's live reload rides on one.
      "/preview": {
        target: apiTarget,
        changeOrigin: true,
        ws: true,
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
