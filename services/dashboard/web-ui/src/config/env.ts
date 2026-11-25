export const env = {
    BFF_HOST: import.meta.env.VITE_BFF_HOST,
    BFF_PORT: import.meta.env.VITE_BFF_PORT,
    BFF_API_PREFIX: import.meta.env.VITE_BFF_API_PREFIX,
} as const;
