import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    // vueDevTools(), // 注释掉此插件以解决按 Ctrl+C 无法退出的问题
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    proxy: {
      '^/api(?:/|$)': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
      '^/images/(random|adaptive)(?:\\?|$)': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
      '^/uploads/bg(?:/|$)': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
    },
  },
})
