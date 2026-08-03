import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import vuetify from 'vite-plugin-vuetify'
import { defineConfig } from 'vite'

// FinvQuant 前端：Vue3 + Vite8 + Vuetify4
// 开发端口 16002；/api 代理到服务端 16001
export default defineConfig({
  plugins: [
    vue(),
    vuetify({ autoImport: true }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    host: '0.0.0.0',
    port: 16002,
    proxy: {
      '/api': {
        target: process.env.FINV_SERVER_URL || 'http://localhost:16001',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
  },
})
