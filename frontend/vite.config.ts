import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    }
  },
  build: {
    emptyOutDir: true,
    outDir: '../dist/public',
    assetsDir: 'assets',
  },
  server: {
    proxy: {
      '/api/ifconfig': {
        target: 'https://ifconfig.me',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/ifconfig/, '')
      }
    }
  }
})
