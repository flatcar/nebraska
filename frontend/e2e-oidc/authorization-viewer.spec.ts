import { expect, test } from '@playwright/test';

import { OIDCHelpers } from './helpers/oidc-helpers';
import { TEST_USERS } from './helpers/test-users';
import { OIDC_TEST_CONFIG } from './test-config';

test.describe('OIDC Authorization - Viewer User', () => {
  let oidcHelpers: OIDCHelpers;

  test.beforeEach(async () => {
    oidcHelpers = new OIDCHelpers();
  });

  test('viewer should have read access to all resources', async ({ request }) => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // Test read access to applications
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps?page=0&perpage=10`,
      viewerToken.token
    );
    expect(appsResult.status).toBe(200);
    expect(appsResult.data.applications).toBeDefined();

    // Test read access to activity logs
    const activityResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/activity?start=${new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString()}&end=${new Date().toISOString()}`,
      viewerToken.token
    );
    expect(activityResult.status).toBe(200);

    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];

      // Test read access to app details, groups, packages
      const appResult = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'GET',
        `${baseUrl}/api/apps/${firstApp.id}`,
        viewerToken.token
      );
      expect(appResult.status).toBe(200);

      const groupsResult = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'GET',
        `${baseUrl}/api/apps/${firstApp.id}/groups`,
        viewerToken.token
      );
      expect(groupsResult.status).toBe(200);

      const packagesResult = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'GET',
        `${baseUrl}/api/apps/${firstApp.id}/packages`,
        viewerToken.token
      );
      expect(packagesResult.status).toBe(200);
    }
  });

  test('viewer should be denied all write operations', async ({ request }) => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // Test POST denial - creating applications
    const newApp = {
      name: 'Viewer Test App',
      product_id: 'viewer-test-app-id',
    };
    const createResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'POST',
      `${baseUrl}/api/apps`,
      viewerToken.token,
      newApp
    );
    expect(createResult.status).toBe(403);

    // Get existing application for PUT/DELETE tests
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps?page=0&perpage=10`,
      viewerToken.token
    );
    expect(appsResult.status).toBe(200);

    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];

      // Test PUT denial - updating applications
      const updateResult = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'PUT',
        `${baseUrl}/api/apps/${firstApp.id}`,
        viewerToken.token,
        { name: 'Updated by viewer' }
      );
      expect(updateResult.status).toBe(403);

      // Test DELETE denial - deleting applications
      const deleteResult = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'DELETE',
        `${baseUrl}/api/apps/${firstApp.id}`,
        viewerToken.token
      );
      expect(deleteResult.status).toBe(403);

      // Test write denial for groups and packages
      const newGroup = {
        name: 'Viewer Test Group',
        policy_max_updates_per_period: 1,
        policy_period_interval: '15 minutes',
      };
      const groupCreateResult = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'POST',
        `${baseUrl}/api/apps/${firstApp.id}/groups`,
        viewerToken.token,
        newGroup
      );
      expect(groupCreateResult.status).toBe(403);

      const newPackage = {
        application_id: firstApp.id,
        arch: 1,
        type: 1,
        url: 'https://example.com/package.gz',
        version: '1.0.0',
      };
      const packageCreateResult = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'POST',
        `${baseUrl}/api/apps/${firstApp.id}/packages`,
        viewerToken.token,
        newPackage
      );
      expect(packageCreateResult.status).toBe(403);
    }
  });

  test('viewer token should only contain viewer role', async () => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);

    // Verify token contains only viewer role
    expect(viewerToken.payload.realm_access.roles).toContain('test_viewer');
    expect(viewerToken.payload.realm_access.roles).not.toContain('test_admin');

    // Verify user claims
    expect(viewerToken.payload.preferred_username).toBe(TEST_USERS.VIEWER.username);
  });

  test('compare viewer vs admin permissions', async ({ request }) => {
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // compare admin vs viewer permissions
    const { adminResult, viewerResult } = await oidcHelpers.testRoleBasedAccess(
      request,
      '/api/apps',
      'GET'
    );

    // both should be able to read
    expect(adminResult.status).toBe(200);
    expect(viewerResult.status).toBe(200);

    // test write operation permissions
    const newApp = {
      name: 'Role Test App',
      product_id: 'role-test-app-id',
    };

    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);

    const adminWrite = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'POST',
      `${baseUrl}/api/apps`,
      adminToken.token,
      newApp
    );
    const viewerWrite = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'POST',
      `${baseUrl}/api/apps`,
      viewerToken.token,
      newApp
    );

    // admin should be authorized
    expect(adminWrite.status).not.toBe(403);
    expect(adminWrite.status).not.toBe(401);

    // viewer should be forbidden
    expect(viewerWrite.status).toBe(403);
  });
});
