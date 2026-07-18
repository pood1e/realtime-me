/// <reference types="vite/client" />
interface ImportMetaEnv {
  readonly VITE_PRIVATE_API_BASE?: string;
  readonly VITE_DEFAULT_RETURN_URL?: string;
  readonly VITE_DRIVE_APP_ORIGIN?: string;
  readonly VITE_BOOKS_APP_ORIGIN?: string;
  readonly VITE_MUSIC_APP_ORIGIN?: string;
  readonly VITE_IMAGES_APP_ORIGIN?: string;
}
