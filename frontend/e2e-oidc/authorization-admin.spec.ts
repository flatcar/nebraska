import { expect, test } from '@playwright/test';

import { OIDCHelpers } from './helpers/oidc-helpers';
import { TEST_USERS } from './helpers/test-users';
import { OIDC_TEST_CONFIG } from './test-config';

test.describe('OIDC Authorization - Admin User', () => {
  let oidcHelpers: OIDCHelpers;

  test.beforeEach(async () => {
    oidcHelpers = new OIDCHelpers();
  });

  test('admin should have full read access to all API endpoints', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // Test access to main endpoints
    const endpoints = ['/api/apps?page=0&perpage=10'];

    for (const endpoint of endpoints) {
      const result = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'GET',
        `${baseUrl}${endpoint}`,
        adminToken.token
      );
      expect(result.status).toBe(200);
      expect(result.data).toBeTruthy();
    }
  });

  test('admin should have write access to applications', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // create a new application
    const newApp = {
      name: 'OIDC Test App',
      product_id: 'oidc-test-app-id',
    };

    const createResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'POST',
      `${baseUrl}/api/apps`,
      adminToken.token,
      newApp
    );

    // admin should not be forbidden
    expect(createResult.status).not.toBe(403);
    expect(createResult.status).not.toBe(401);
  });

  test('admin should have access to application details', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // First get list of applications
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps`,
      adminToken.token
    );
    expect(appsResult.status).toBe(200);

    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];

      // Test access to specific application
      const appResult = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'GET',
        `${baseUrl}/api/apps/${firstApp.id}`,
        adminToken.token
      );
      expect(appResult.status).toBe(200);
      expect(appResult.data).toBeTruthy();
    }
  });

  test('admin should have access to groups management', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // Get applications first
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps`,
      adminToken.token
    );
    expect(appsResult.status).toBe(200);

    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];

      // admin should have access to groups
      const groupsResult = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'GET',
        `${baseUrl}/api/apps/${firstApp.id}/groups`,
        adminToken.token
      );
      expect(groupsResult.status).toBe(200);
    }
  });

  test('admin should have access to packages management', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // Get applications first
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps`,
      adminToken.token
    );
    expect(appsResult.status).toBe(200);

    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];

      // admin should have access to packages
      const packagesResult = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'GET',
        `${baseUrl}/api/apps/${firstApp.id}/packages`,
        adminToken.token
      );
      expect(packagesResult.status).toBe(200);
    }
  });

  test('admin should have access to instances data', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // Get applications first
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps?page=0&perpage=10`,
      adminToken.token
    );
    expect(appsResult.status).toBe(200);

    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];

      // Get groups for the application
      const groupsResult = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'GET',
        `${baseUrl}/api/apps/${firstApp.id}/groups`,
        adminToken.token
      );

      if (groupsResult.status === 200 && groupsResult.data?.groups?.length > 0) {
        const firstGroup = groupsResult.data.groups[0];

        // Test instances endpoint - admin should have access
        const instancesResult = await oidcHelpers.makeAuthenticatedRequest(
          request,
          'GET',
          `${baseUrl}/api/apps/${firstApp.id}/groups/${firstGroup.id}/instances`,
          adminToken.token
        );

        // Admin should not be forbidden (403) or unauthorized (401)
        expect(instancesResult.status).not.toBe(403);
        expect(instancesResult.status).not.toBe(401);
      }
    }

    // If no test data exists, we've still verified admin can access apps/groups
  });

  test('admin should have access to activity logs', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // admin should have access to activity logs
    const activityResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/activity?start=${new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString()}&end=${new Date().toISOString()}`,
      adminToken.token
    );
    expect(activityResult.status).toBe(200);
  });

  test('admin role should include viewer permissions', async ({ request }) => {
    // Admin users should have composite roles that include viewer permissions
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);

    // Verify token contains both admin and viewer roles
    expect(adminToken.payload.realm_access.roles).toContain('test_admin');
    expect(adminToken.payload.realm_access.roles).toContain('test_viewer');

    // admin should be able to perform viewer-level operations
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps`,
      adminToken.token
    );
    expect(appsResult.status).toBe(200);
  });

  test('admin should be able to create applications', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // Test admin can create applications
    const newApp = {
      name: 'Admin Test App',
      product_id: 'admin-test-app-id',
    };

    const result = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'POST',
      `${baseUrl}/api/apps`,
      adminToken.token,
      newApp
    );

    // admin should not be forbidden or unauthorized
    expect(result.status).not.toBe(403);
    expect(result.status).not.toBe(401);
  });
});
