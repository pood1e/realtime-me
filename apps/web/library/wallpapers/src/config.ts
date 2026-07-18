import { apiBaseUrl } from "@realtime-me/library-web";

export const API_BASE = apiBaseUrl(import.meta.env.VITE_PUBLIC_API_BASE, "http://localhost:8080");
