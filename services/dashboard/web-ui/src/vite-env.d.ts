/// <reference types="vite/client" />
//
interface ImportMetaEnv {
    readonly VITE_BFF_HOST: string;
    readonly VITE_BFF_PORT: string;
    readonly VITE_BFF_API_PREFIX: string;
}

interface ImportMeta {
    readonly env: ImportMetaEnv;
}
