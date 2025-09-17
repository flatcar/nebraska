// Shared OIDC test configuration
// This file provides centralized configuration for all OIDC tests
export const OIDC_TEST_CONFIG = {
  nebraska: {
    baseURL: process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003',
  },
  keycloak: {
    baseURL: process.env.CI ? 'http://127.0.0.1:8063' : 'http://localhost:8063',
    realm: 'test',
    clientId: 'nebraska-test',
  },
  database: {
    host: process.env.CI ? '127.0.0.1' : 'localhost',
    port: 8001,
    name: 'nebraska_tests',
    user: 'postgres',
    password: 'nebraska',
  },
};
