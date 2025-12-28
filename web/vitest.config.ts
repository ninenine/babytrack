import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.test.{ts,tsx}'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      include: ['src/**/*.{ts,tsx}'],
      exclude: [
        'node_modules/',
        'src/test/',
        '**/*.d.ts',
        '**/*.test.{ts,tsx}',
        'src/main.tsx',
        'src/vite-env.d.ts',
        'src/components/**', // React components - better for integration tests
        'src/app/**', // app config (routes, providers)
        'src/App.tsx', // root component
        'src/pages/**', // pages - better for E2E tests
        'src/hooks/index.ts', // barrel exports
        'src/hooks/queries/**', // query hooks - require complex mocking
        'src/hooks/use-notifications.ts', // SSE hook - integration test
        'src/hooks/use-sync.ts', // sync hook - uses tested sync.ts
        'src/lib/query-client.ts', // query client config
      ],
      thresholds: {
        global: {
          branches: 90,
          functions: 90,
          lines: 90,
          statements: 90,
        },
      },
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})
