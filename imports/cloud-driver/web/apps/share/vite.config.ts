import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import { cloudflarePagesHeaders } from "../../build/cloudflare-pages";
import { DEFAULT_SHARE_API_BASE } from "./src/config";

export default defineConfig({
  plugins: [react(), tailwindcss(), cloudflarePagesHeaders("VITE_SHARE_API_BASE", DEFAULT_SHARE_API_BASE)],
  build: { target: "es2022" },
});
