import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

// https://vite.dev/config/
export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: 'src/setupTests.ts',
    include: ['src/**/*.{spec,test}.{js,jsx,ts,tsx}'],
    coverage: {
      provider: "istanbul",
      reporter: ['text', 'text-summary'],
      include: ['src/**/*.{js,jsx,ts,tsx}'],
      exclude: [
        'src/setupTests.ts',
        'e2e/',
        'node_modules/',
        'build/',
        'src/**/*.stories*.{js,jsx,ts,tsx}',
        'src/components/**/index.tsx',
      ],
      thresholds: {
        lines: 23.6,
        functions: 21.09,
        statements: 23.57,
        branches: 17,
      }
    },
  },
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
      '/config': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
    },
  },
});
