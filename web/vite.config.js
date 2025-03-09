import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api/service': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => {
          const matches = path.match(/\/api\/service\/([^\/]+)\/(.+)/)
          if (matches) {
            return `/service-control?service_name=${matches[1]}&action=${matches[2]}`
          }
          return path
        }
      }
    }
  }
})