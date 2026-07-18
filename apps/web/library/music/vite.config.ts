import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { cloudflarePagesHeaders } from "../../build/cloudflare-pages";

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    cloudflarePagesHeaders("VITE_PRIVATE_API_BASE", "http://localhost:8080", {
      script: ["https://sdk.scdn.co"],
      image: [
        "https://*.gtimg.cn",
        "https://*.qq.com",
        "https://*.music.126.net",
        "https://*.music.163.com",
        "https://*.scdn.co",
      ],
      media: [
        "https://*.qqmusic.qq.com",
        "https://*.music.126.net",
        "https://*.music.163.com",
        "https://*.scdn.co",
      ],
      connect: [
        "https://api.spotify.com",
        "https://*.spotify.com",
        "https://*.scdn.co",
        "wss://*.spotify.com",
      ],
    }),
  ],
  server: { port: 5176 },
  build: { target: "es2022" },
});
