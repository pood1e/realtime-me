import { apiBaseUrl } from "@cloud-drive/shared";

export const API_BASE = apiBaseUrl(
  import.meta.env.VITE_PUBLIC_API_BASE,
  "http://localhost:8080",
);
