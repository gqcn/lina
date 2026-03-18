import { readFileSync } from 'node:fs';
import { join } from 'node:path';

import { defineConfig } from '@vben/vite-config';

// Cache the HTML content to avoid repeated synchronous file reads
let cachedApidocsHtml: string | undefined;

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
            target: 'http://localhost:8080',
            bypass(_req, res) {
              // Serve the static HTML file directly, bypassing Vite's SPA fallback
              if (!cachedApidocsHtml) {
                const filePath = join(
                  import.meta.dirname,
                  'public/stoplight/apidocs.html',
                );
                cachedApidocsHtml = readFileSync(filePath, 'utf-8');
              }
              res.setHeader('Content-Type', 'text/html; charset=utf-8');
              res.end(cachedApidocsHtml);
              // Return false to prevent proxy from connecting to the target
              return false;
            },
          },
        },
        watch: {
          // Exclude directories that don't need HMR watching
          ignored: [
            '**/public/stoplight/**',
            '**/node_modules/**',
            '**/dist/**',
            '**/.vite/**',
          ],
        },
      },
    },
  };
});
