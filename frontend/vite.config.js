import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

const apiPrefixes = [
  '/config',
  '/categories',
  '/currency',
  '/startdate',
  '/expense',
  '/expenses',
  '/recurring-expense',
  '/recurring-expenses',
  '/import',
  '/export',
  '/login',
  '/logout',
  '/api'
];

const proxy = apiPrefixes.reduce((acc, prefix) => {
  acc[prefix] = {
    target: 'http://localhost:9080',
    changeOrigin: true,
  };
  return acc;
}, {});

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy,
  },
  build: {
    outDir: '../internal/web/dist',
    emptyOutDir: true,
    sourcemap: false
  }
});
