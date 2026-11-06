import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

const backendUrl = process.env.MUD_BACKEND_URL ?? 'http://localhost:7272'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api/websocket': {
        target: backendUrl,
        ws: true,
        changeOrigin: true
      },
      '/api': {
        target: backendUrl,
        changeOrigin: true
      }
    }
  }
})
