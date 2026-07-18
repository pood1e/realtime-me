import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import { cloudflarePagesHeaders } from "../../build/cloudflare-pages";

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    cloudflarePagesHeaders("VITE_PRIVATE_API_BASE", "http://localhost:8080"),
  ],
  server: { port: 5175 },
  build: { target: "es2022" },
});
