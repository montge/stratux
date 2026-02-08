import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],

  // Build configuration for Stratux embedded device
  build: {
    // Output to dist folder (will be copied to /opt/stratux/www/)
    outDir: 'dist',

    // Generate source maps for debugging
    sourcemap: false,

    // Minify for smaller bundle
    minify: 'esbuild',

    // Target modern browsers (Raspberry Pi uses Chromium)
    target: 'es2020',

    // Chunk size warning limit
    chunkSizeWarningLimit: 500,

    rollupOptions: {
      output: {
        // Keep asset names predictable for caching
        assetFileNames: 'assets/[name]-[hash][extname]',
        chunkFileNames: 'assets/[name]-[hash].js',
        entryFileNames: 'assets/[name]-[hash].js',
      }
    }
  },

  // Development server configuration
  server: {
    port: 3000,
    // Proxy API requests to actual Stratux device during development
    proxy: {
      '/getStatus': 'http://192.168.10.1',
      '/getSettings': 'http://192.168.10.1',
      '/setSettings': 'http://192.168.10.1',
      '/getSituation': 'http://192.168.10.1',
      '/getTowers': 'http://192.168.10.1',
      '/status': {
        target: 'ws://192.168.10.1',
        ws: true
      },
      '/traffic': {
        target: 'ws://192.168.10.1',
        ws: true
      },
      '/situation': {
        target: 'ws://192.168.10.1',
        ws: true
      },
      '/weather': {
        target: 'ws://192.168.10.1',
        ws: true
      },
      '/radar': {
        target: 'ws://192.168.10.1',
        ws: true
      }
    }
  },

  // Preview server (for testing production build)
  preview: {
    port: 4173
  }
})
