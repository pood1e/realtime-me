/// <reference types="vite/client" />

// The /apps launcher points at one site — a public commons with a music, a books and a
// wallpapers page — which lives in its own project. Its address is build-time configuration
// rather than anything this page can derive, and if it is unset the launcher says so rather
// than offering links that go nowhere.
interface ImportMetaEnv {
  readonly VITE_COMMONS_APP_URL?: string;
}
