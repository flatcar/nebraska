import { expect, test } from '@playwright/test';

import { OIDCHelpers } from './helpers/oidc-helpers';
import { TEST_USERS } from './helpers/test-users';

test.describe('OIDC API Endpoint Protection', () => {
  let oidcHelpers: OIDCHelpers;

  test.beforeEach(async () => {
    oidcHelpers = new OIDCHelpers();
  });

  test('should protect all API endpoints from unauthenticated access', async ({ request }) => {
    // List of API endpoints that should be protected
    const protectedEndpoints = ['/api/apps', '/api/activity'];

    for (const endpoint of protectedEndpoints) {
      // Test without any authorization header
      const noAuthResponse = await request.get(endpoint);
      expect(noAuthResponse.status()).toBe(401);

      // Test with invalid authorization header
      const invalidAuthResponse = await request.get(endpoint, {
        headers: { Authorization: 'Invalid header format' },
      });
      expect(invalidAuthResponse.status()).toBe(401);
    }
  });

  test('should allow access to unprotected config endpoint', async ({ request }) => {
    // /config endpoint should be accessible without authentication for frontend bootstrap
    const configResponse = await request.get('/config');
    expect(configResponse.status()).toBe(200);

    const configData = await configResponse.json();
    expect(configData).toHaveProperty('auth_mode', 'oidc');
    expect(configData).toHaveProperty('oidc_client_id');
    expect(configData).toHaveProperty('oidc_issuer_url');
  });

  test('should validate token on every API request', async ({ request }) => {
    const validToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const invalidToken = 'invalid-token-12345';

    // Define endpoints with their required parameters
    const endpoints = [
      '/api/apps?page=0&perpage=10',
      `/api/activity?start=${new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString()}&end=${new Date().toISOString()}`,
    ];

    for (const endpoint of endpoints) {
      // Valid token should work
      const validResult = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'GET',
        endpoint,
        validToken.token
      );
      expect(validResult.status).toBe(200);

      // Invalid token should be rejected
      const invalidResult = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'GET',
        endpoint,
        invalidToken
      );
      expect(invalidResult.status).toBe(401);
    }
  });

  test('should enforce role-based access control on write operations', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);

    // Test application creation
    const newApp = {
      name: 'RBAC Test App',
      product_id: 'rbac-test-id',
      description: 'Testing role-based access control',
    };

    // Admin should be allowed to create (or get business logic error, not auth error)
    const adminCreateResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'POST',
      '/api/apps',
      adminToken.token,
      newApp
    );
    expect(adminCreateResult.status).not.toBe(403); // Not forbidden
    expect(adminCreateResult.status).not.toBe(401); // Not unauthorized

    // Viewer should be forbidden from creating
    const viewerCreateResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'POST',
      '/api/apps',
      viewerToken.token,
      newApp
    );
    expect(viewerCreateResult.status).toBe(403); // Should be forbidden
  });

  test('should handle HEAD requests properly', async ({ request }) => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);

    // HEAD requests should be allowed for viewers (read-only operation)
    const headResponse = await request.head('/api/apps', {
      headers: { Authorization: `Bearer ${viewerToken.token}` },
    });

    // HEAD should be treated as read operation, allowed for viewers
    expect(headResponse.status()).toBe(200);
  });

  test('should properly handle OPTIONS requests', async ({ request }) => {
    // OPTIONS requests are typically for CORS preflight
    const optionsResponse = await request.fetch('/api/apps', {
      method: 'OPTIONS',
    });

    // OPTIONS should be handled appropriately (200 or 204)
    expect([200, 204, 405]).toContain(optionsResponse.status());
  });

  test('should protect nested API endpoints', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);

    // First get an application to work with
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      '/api/apps',
      adminToken.token
    );
    expect(appsResult.status).toBe(200);

    if (appsResult.data?.applications?.length > 0) {
      const appId = appsResult.data.applications[0].id;

      // Test nested endpoints
      const nestedEndpoints = [
        `/api/apps/${appId}`,
        `/api/apps/${appId}/groups`,
        `/api/apps/${appId}/packages`,
      ];

      for (const endpoint of nestedEndpoints) {
        // Test without auth
        const noAuthResponse = await request.get(endpoint);
        expect(noAuthResponse.status()).toBe(401);

        // Test with valid auth
        const authResult = await oidcHelpers.makeAuthenticatedRequest(
          request,
          'GET',
          endpoint,
          adminToken.token
        );
        expect(authResult.status).toBe(200);
      }
    }
  });
});
