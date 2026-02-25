import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'

const apiTarget = process.env.API_TARGET || 'http://localhost:5678'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    host: '0.0.0.0',
    port: 3012,
    allowedHosts: ['.trycloudflare.com', '.loca.lt'],
    proxy: {
      '/api': {
        target: apiTarget,
        changeOrigin: true
      },
      '/static': {
        target: apiTarget,
        changeOrigin: true
      }
    }
  }
})
