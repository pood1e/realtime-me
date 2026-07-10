/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_SHARE_API_BASE?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
