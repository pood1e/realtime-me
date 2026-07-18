import { apiBaseUrl } from "./api";

export type PrivateAppId = "drive" | "books" | "music" | "images";
export type AppLinks = Partial<
  Record<PrivateAppId | "wallpapers" | "share", string>
>;

export type PrivateAppEnvironment = Readonly<{
  VITE_PRIVATE_API_BASE?: string;
  VITE_AUTH_APP_ORIGIN?: string;
  VITE_DRIVE_APP_ORIGIN?: string;
  VITE_BOOKS_APP_ORIGIN?: string;
  VITE_MUSIC_APP_ORIGIN?: string;
  VITE_IMAGES_APP_ORIGIN?: string;
  VITE_WALLPAPERS_APP_ORIGIN?: string;
  VITE_SHARE_APP_ORIGIN?: string;
}>;

export function privateAppConfiguration(environment: PrivateAppEnvironment) {
  return {
    apiBase: apiBaseUrl(
      environment.VITE_PRIVATE_API_BASE,
      "http://localhost:8080",
    ),
    authOrigin:
      environment.VITE_AUTH_APP_ORIGIN?.trim() || "http://localhost:5173",
    links: {
      drive: environment.VITE_DRIVE_APP_ORIGIN || "http://localhost:5174",
      books: environment.VITE_BOOKS_APP_ORIGIN || "http://localhost:5175",
      music: environment.VITE_MUSIC_APP_ORIGIN || "http://localhost:5176",
      images: environment.VITE_IMAGES_APP_ORIGIN || "http://localhost:5177",
      wallpapers:
        environment.VITE_WALLPAPERS_APP_ORIGIN || "http://localhost:5178",
      share: environment.VITE_SHARE_APP_ORIGIN || "http://localhost:5179",
    } satisfies AppLinks,
  } as const;
}
