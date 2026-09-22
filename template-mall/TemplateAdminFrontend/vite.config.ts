import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3001,
    strictPort: true,
    proxy: {
      // OAuth 回调：http://localhost:3001/api/auth/callback → BFF /auth/callback
      '/api/auth': {
        target: 'http://localhost:8081',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/auth/, '/auth'),
      },
      // 管理 API：/api/admin/* → BFF /api/admin/*
      '/api/admin': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  },
  preview: {
    port: 3001,
    strictPort: true,
  },
});
