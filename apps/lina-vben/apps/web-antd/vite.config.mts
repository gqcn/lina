import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
  return {
    application: {},
    vite: {
      server: {
        proxy: {
          '/api': {
            changeOrigin: true,
            // Forward /api/* to backend at localhost:8080/api/*
            target: 'http://localhost:8080',
            ws: true,
          },
        },
      },
    },
  };
});
