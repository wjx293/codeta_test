/// <reference types="vite/client" />

declare module 'ali-oss';

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string;
  readonly VITE_AUTH_MODE: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
