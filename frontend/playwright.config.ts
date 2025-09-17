import { defineConfig } from '@playwright/test';
import { loadEnv } from 'vite';

import { chromeWithConsistentRendering } from './playwright.shared.config';


export const ENV_DIR = './';
Object.assign(process.env, loadEnv('', ENV_DIR));

/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
  testDir: './e2e',
  /* Run tests in files in parallel */
  fullyParallel: true,
  /* Fail the build on CI if you accidentally left test.only in the source code. */
  forbidOnly: !!process.env.CI,
  /* Retry on CI only */
  retries: process.env.CI ? 2 : 0,
  /* Run tests in parallel with limited workers to reduce contention */
  workers: undefined,
  /* Reporter to use. See https://playwright.dev/docs/test-reporters */
  reporter: [['line'], ['html']],
  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  timeout: 50_000, 
  globalTimeout: process.env.CI ? 600_000 : undefined, // 10 minutes total for CI
  use: {
    baseURL: process.env.CI ? 'http://127.0.0.1:8002' : 'http://localhost:3000',

    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    trace: 'on-first-retry',
  },

  /* Configure projects for major browsers */
  projects: [
    {
      name: 'setup db',
      testMatch: /global\.setup\.ts/,
      teardown: 'cleanup db',
    },

    {
      name: 'cleanup db',
      testMatch: /global\.teardown\.ts/,
    },

    {
      name: 'chrome',
      use: chromeWithConsistentRendering,
      dependencies: ['setup db'],
    },
  ],

  /* Run your local dev server before starting the tests */
  webServer: process.env.CI ? {
    command: 'cd ../backend && docker compose -f docker-compose.test.yaml up --build --force-recreate && docker ps -a',
    url: 'http://127.0.0.1:8002',
    reuseExistingServer: false,
    timeout: 200_000,
  } : {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: true,
    timeout: 120_000,
  },
  expect: {
    toHaveScreenshot: {
      stylePath: './e2e/mask-and-fix-dynamic-parts.css',
    }
  }
});
