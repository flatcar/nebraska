import { test as teardown } from '@playwright/test';

teardown('cleanup OIDC test environment', async () => {
  console.log('Cleaning up OIDC test environment...');

  // In CI, docker compose will handle cleanup
  // For local development, we could add cleanup logic here if needed

  console.log('OIDC test environment cleanup completed');
});
