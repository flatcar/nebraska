import { defineConfig, devices } from '@playwright/test';
import { loadEnv } from 'vite';


export const ENV_DIR = './';
Object.assign(process.env, loadEnv('', ENV_DIR));

console.log('CI=', process.env.CI);

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
  /* Opt out of parallel tests on CI. */
  workers: process.env.CI ? 1 : undefined,
  /* Reporter to use. See https://playwright.dev/docs/test-reporters */
  reporter: 'html',
  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  timeout: 120_000,
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
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
      dependencies: ['setup db'],
    },
  ],

  /* Run your local dev server before starting the tests */
  webServer: process.env.CI ? {
    command: 'cd ../backend && docker compose -f docker-compose.test.yaml up && docker ps -a',
    url: 'http://127.0.0.1:8002', // Replace with the URL of your service
    reuseExistingServer: false,
    timeout: 200_000,
  } : undefined,
  expect: {
    toHaveScreenshot: {
      stylePath: './e2e/mask-and-fix-dynamic-parts.css',
    }
  }
});
