import { defineConfig, devices } from '@playwright/test';
import { loadEnv } from 'vite';

export const ENV_DIR = './';
Object.assign(process.env, loadEnv('', ENV_DIR));

console.log('OIDC CI=', process.env.CI);

/**
 * See https://playwright.dev/docs/test-configuration.
 * 
 * OIDC-specific test configuration that runs separately from main tests.
 * Uses different ports and includes Keycloak for authentication testing.
 */
export default defineConfig({
  testDir: './e2e-oidc',
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
    baseURL: process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003',

    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    trace: 'on-first-retry',
  },

  /* Configure projects for major browsers */
  projects: [
    {
      name: 'oidc-setup',
      testMatch: /oidc-global\.setup\.ts/,
      teardown: 'oidc-cleanup',
    },

    {
      name: 'oidc-cleanup',
      testMatch: /oidc-global\.teardown\.ts/,
    },

    {
      name: 'oidc-firefox',
      use: { ...devices['Desktop Firefox'] },
      dependencies: ['oidc-setup'],
    },
  ],

  /* Run your local dev server before starting the tests */
  webServer: process.env.CI ? {
    command: 'cd ../backend && docker compose -f docker-compose.base.yaml -f docker-compose.oidc-test.yaml up --build && docker ps -a',
    url: 'http://127.0.0.1:8003', // Nebraska backend with OIDC
    reuseExistingServer: false,
    timeout: 300_000, // 5 minutes for Keycloak + Nebraska startup
  } : undefined,

  expect: {
    toHaveScreenshot: {
      stylePath: './e2e-oidc/mask-oidc-dynamic-parts.css',
    }
  }
});
