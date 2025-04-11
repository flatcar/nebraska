/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_PRIMARY_COLOR: string;
  readonly VITE_PROJECT_NAME: string;
  readonly VITE_APPBAR_BG: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
