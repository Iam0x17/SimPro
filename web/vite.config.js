import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  base: '/admin/',
  server: {
    proxy: {
      '/api/service': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})