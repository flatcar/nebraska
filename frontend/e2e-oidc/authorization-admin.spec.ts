import { test, expect } from '@playwright/test';
import { OIDCHelpers } from './helpers/oidc-helpers';
import { TEST_USERS } from './helpers/test-users';

test.describe('OIDC Authorization - Admin User', () => {
  let oidcHelpers: OIDCHelpers;

  test.beforeEach(async () => {
    oidcHelpers = new OIDCHelpers();
  });

  test('admin should have full read access to all API endpoints', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Test access to main endpoints
    const endpoints = [
      '/api/apps?page=0&perpage=10',
    ];
    
    for (const endpoint of endpoints) {
      const result = await oidcHelpers.makeAuthenticatedRequest(
        request, 'GET', `${baseUrl}${endpoint}`, adminToken.token
      );
      expect(result.status).toBe(200);
      expect(result.data).toBeTruthy();
    }
  });

  test('admin should have write access to applications', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Test creating a new application (if endpoint supports it)
    const newApp = {
      name: 'OIDC Test App',
      product_id: 'oidc-test-app-id',
      description: 'Test application for OIDC E2E tests'
    };
    
    const createResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'POST', `${baseUrl}/api/apps`, adminToken.token, newApp
    );
    
    // Should either succeed (201) or fail due to business logic (400/409/500), not authorization (403)
    expect([200, 201, 400, 409, 500]).toContain(createResult.status);
    expect(createResult.status).not.toBe(403);
  });

  test('admin should have access to application details', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // First get list of applications
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps`, adminToken.token
    );
    expect(appsResult.status).toBe(200);
    
    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];
      
      // Test access to specific application
      const appResult = await oidcHelpers.makeAuthenticatedRequest(
        request, 'GET', `${baseUrl}/api/apps/${firstApp.id}`, adminToken.token
      );
      expect(appResult.status).toBe(200);
      expect(appResult.data).toBeTruthy();
    }
  });

  test('admin should have access to groups management', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Get applications first
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps`, adminToken.token
    );
    expect(appsResult.status).toBe(200);
    
    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];
      
      // Test access to groups for this application
      const groupsResult = await oidcHelpers.makeAuthenticatedRequest(
        request, 'GET', `${baseUrl}/api/apps/${firstApp.id}/groups`, adminToken.token
      );
      expect(groupsResult.status).toBe(200);
    }
  });

  test('admin should have access to packages management', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Get applications first
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps`, adminToken.token
    );
    expect(appsResult.status).toBe(200);
    
    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];
      
      // Test access to packages for this application
      const packagesResult = await oidcHelpers.makeAuthenticatedRequest(
        request, 'GET', `${baseUrl}/api/apps/${firstApp.id}/packages?page=0&perpage=10`, adminToken.token
      );
      expect(packagesResult.status).toBe(200);
    }
  });

  test('admin should have access to instances data', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Get applications first
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps?page=0&perpage=10`, adminToken.token
    );
    expect(appsResult.status).toBe(200);
    
    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];
      
      if (firstApp.groups?.length > 0) {
        const firstGroup = firstApp.groups[0];
        
        // Test access to instances for this group
        const instancesResult = await oidcHelpers.makeAuthenticatedRequest(
          request, 'GET', `${baseUrl}/api/apps/${firstApp.id}/groups/${firstGroup.id}/instances?status=0&sort=2&sortOrder=0&page=1&perpage=10&duration=30d`, adminToken.token
        );
        expect(instancesResult.status).toBe(200);
      }
    }
  });

  test('admin should have access to activity logs', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Test access to activity endpoint
    const activityResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/activity?start=${new Date(Date.now() - 24*60*60*1000).toISOString()}&end=${new Date().toISOString()}`, adminToken.token
    );
    expect(activityResult.status).toBe(200);
  });

  test('admin role should include viewer permissions', async ({ request }) => {
    // Admin users should have composite roles that include viewer permissions
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    
    // Verify token contains both admin and viewer roles
    expect(adminToken.payload.realm_access.roles).toContain('test_admin');
    expect(adminToken.payload.realm_access.roles).toContain('test_viewer');
    
    // Test that admin can perform viewer-level operations
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps?page=0&perpage=10`, adminToken.token
    );
    expect(appsResult.status).toBe(200);
  });

  test('admin should be able to perform write operations', async ({ request }) => {
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Test that admin can make POST requests (create operations)
    // Note: Actual creation might fail due to business logic, but should not fail due to authorization
    
    const testEndpoints = [
      { method: 'POST' as const, endpoint: '/api/apps', data: { name: 'test', product_id: 'test' } },
    ];
    
    for (const { method, endpoint, data } of testEndpoints) {
      const result = await oidcHelpers.makeAuthenticatedRequest(
        request, method, `${baseUrl}${endpoint}`, adminToken.token, data
      );
      
      // Should not fail with 403 (Forbidden) - admin should have permission
      expect(result.status).not.toBe(403);
      
      // Should not fail with 401 (Unauthorized) - token should be valid
      expect(result.status).not.toBe(401);
    }
  });
});