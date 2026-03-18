import { readFileSync } from 'node:fs';
import { join } from 'node:path';

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
          '/stoplight/apidocs.html': {
            target: 'http://localhost:5666',
            bypass(_req, res) {
              // Serve the static HTML file directly, bypassing Vite's SPA fallback
              const filePath = join(
                import.meta.dirname,
                'public/stoplight/apidocs.html',
              );
              const content = readFileSync(filePath, 'utf-8');
              res.setHeader('Content-Type', 'text/html; charset=utf-8');
              res.end(content);
            },
          },
        },
      },
    },
  };
});
