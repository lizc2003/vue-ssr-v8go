import { defineConfig } from 'vite'
import config from './vite.config'

export default defineConfig({
  ...config,
  build: {
    emptyOutDir: true,
    outDir: '../dist/server',
    ssr: 'src/entry-server.ts',
    copyPublicDir: false,
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
