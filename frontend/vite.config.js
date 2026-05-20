import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: parseInt(process.env.PORT) || 5173,
    strictPort: false,
    proxy: {
      '/teams':   'http://localhost:8080',
      '/simulate': 'http://localhost:8080',
    },
  },
})
