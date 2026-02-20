import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import { fileURLToPath } from 'url'
import { dirname, resolve } from 'path'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': resolve(__dirname, './src'),
    },
  },
  // @ts-expect-error vitest extends vite config
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/ws/terminal': {
        target: 'ws://localhost:8080',
        ws: true,
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
  build: {
    target: 'es2015',
    minify: 'terser',
    sourcemap: false,
    chunkSizeWarningLimit: 520,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (id.includes('react-dom') || id.includes('react-router-dom') || id.includes('react-i18next')) return 'react-vendor'
            if (id.includes('/react/')) return 'react-vendor'
            if (id.includes('@xyflow')) return 'flow-vendor'
            if (id.includes('lucide-react')) return 'icons'
            if (id.includes('@radix-ui')) return 'radix-ui'
            if (id.includes('react-hook-form') || id.includes('@hookform') || id.includes('/zod/')) return 'form-vendor'
            if (id.includes('i18next')) return 'i18n-vendor'
            if (id.includes('@monaco-editor') || id.includes('monaco-editor')) return 'monaco'
            if (id.includes('zustand') || id.includes('axios') || id.includes('lodash') || id.includes('date-fns')) return 'data-vendor'
          }
        },
      },
    },
  },
  optimizeDeps: {
    include: ['react', 'react-dom', '@xyflow/react', 'zustand'],
  },
})
