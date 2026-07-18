import { privateAppConfiguration } from "@realtime-me/library-web";

const configuration = privateAppConfiguration(import.meta.env);

export const API_BASE = configuration.apiBase;
export const PRIVATE_APP_ORIGINS = new Set(
  [
    configuration.links.drive,
    configuration.links.books,
    configuration.links.music,
    configuration.links.images,
  ].filter((origin): origin is string => Boolean(origin)),
);
export const DEFAULT_RETURN_URL =
  import.meta.env.VITE_DEFAULT_RETURN_URL?.trim() || configuration.links.books!;
