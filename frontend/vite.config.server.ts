import { defineConfig } from 'vite'
import config from './vite.config'

export default defineConfig({
  ...config,
  build: {
    emptyOutDir: true,
    outDir: '../dist/server',
    rollupOptions: {
      output: {
        format: 'iife',
        entryFileNames: 'server.js',
      },
    },
  },
  ssr: {
    target: 'webworker',
    noExternal: true,
  },
})
