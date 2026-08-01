import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import path from 'node:path'

// Vite config for the certsync admin SPA.
//
// The build output goes directly into dist so the Go `//go:embed all:dist`
// in src/frontend/embed.go picks it up without any copy step. We use `base: '/'`
// because the server serves the SPA at the site root.
//
// In dev mode, `/front` and `/api` are proxied to the locally-running Go
// server (default port 8081, matching init/config/server.yml) so the frontend
// can hit the real backend with same-origin requests.
const serverTarget = process.env.CERTSYNC_DEV_SERVER ?? 'http://localhost:8081'

export default defineConfig({
  base: '/',
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  build: {
    outDir: path.resolve(__dirname, 'dist'),
    emptyOutDir: true,
    target: 'es2018',
    sourcemap: false,
    // Element Plus 是全量引入的（main.ts 里的 app.use(ElementPlus)），若不拆分，
    // 入口 chunk 会把它和 Vue/Pinia/vue-i18n 打在一起，超过 500 kB 的默认告警阈值。
    // 这里把稳定的 vendor 依赖拆成各自可缓存的 chunk，并提高阈值，避免 Element
    // Plus chunk 单独触发告警。
    chunkSizeWarningLimit: 1500,
    rollupOptions: {
      output: {
        manualChunks: {
          vue: ['vue', 'vue-router', 'pinia'],
          'element-plus': ['element-plus', '@element-plus/icons-vue']
        }
      }
    }
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/front': {
        target: serverTarget,
        changeOrigin: true
      },
      '/api': {
        target: serverTarget,
        changeOrigin: true
      }
    }
  }
})
