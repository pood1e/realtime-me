import { fileURLToPath, URL } from "node:url";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    proxy: {
      "/realtime.me.status.v1.StatusService/GetPublicStatus": "http://localhost:18080",
      "/realtime.me.site.v1": "http://localhost:18080",
      "/realtime.me.library.wallpapers.v1.WallpaperPublicService": "http://localhost:8080",
      "/realtime.me.library.drive.v1.ShareService": "http://localhost:8080",
      "/v1": "http://localhost:8080",
      "/i": "http://localhost:8080",
    },
  },
});
