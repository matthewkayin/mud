import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    // The server embeds the build when compiled with -tags embed (see backend/web)
    outDir: '../backend/web/dist',
    emptyOutDir: true
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api/websocket': {
        target: process.env.MUD_BACKEND_URL ?? 'http://localhost:7272',
        ws: true,
        changeOrigin: true
      },
      '/api': {
        target: process.env.MUD_BACKEND_URL ?? 'http://localhost:7272',
        changeOrigin: true
      }
    }
  }
})
