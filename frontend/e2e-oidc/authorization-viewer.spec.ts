import { test, expect } from '@playwright/test';
import { OIDCHelpers } from './helpers/oidc-helpers';
import { TEST_USERS } from './helpers/test-users';

test.describe('OIDC Authorization - Viewer User', () => {
  let oidcHelpers: OIDCHelpers;

  test.beforeEach(async () => {
    oidcHelpers = new OIDCHelpers();
  });

  test('viewer should have read access to applications', async ({ request }) => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Test read access to applications
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps?page=0&perpage=10`, viewerToken.token
    );
    expect(appsResult.status).toBe(200);
    expect(appsResult.data.applications).toBeDefined();
  });

  test('viewer should have read access to config', async ({ request }) => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Test read access to config
    const configResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/config`, viewerToken.token
    );
    expect(configResult.status).toBe(200);
  });

  test('viewer should be denied write access to applications', async ({ request }) => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // viewer cannot create applications
    const newApp = {
      name: 'Viewer Test App'
    };
    
    const createResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'POST', `${baseUrl}/api/apps`, viewerToken.token, newApp
    );
    
    // viewer should be forbidden for write operations
    expect(createResult.status).toBe(403);
  });

  test('viewer should be denied PUT operations', async ({ request }) => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // First get an existing application
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps?page=0&perpage=10`, viewerToken.token
    );
    expect(appsResult.status).toBe(200);
    
    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];
      
      // try to update the application
      const updateData = {
        name: 'Updated by viewer'
      };
      
      const updateResult = await oidcHelpers.makeAuthenticatedRequest(
        request, 'PUT', `${baseUrl}/api/apps/${firstApp.id}`, viewerToken.token, updateData
      );
      
      // viewer should be forbidden for write operations
      expect(updateResult.status).toBe(403);
    }
  });

  test('viewer should be denied DELETE operations', async ({ request }) => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // First get an existing application
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps?page=0&perpage=10`, viewerToken.token
    );
    expect(appsResult.status).toBe(200);
    
    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];
      
      // Try to delete the application
      const deleteResult = await oidcHelpers.makeAuthenticatedRequest(
        request, 'DELETE', `${baseUrl}/api/apps/${firstApp.id}`, viewerToken.token
      );
      
      // Should be forbidden for write operations
      expect(deleteResult.status).toBe(403);
    }
  });

  test('viewer should have read access to application details', async ({ request }) => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // First get list of applications
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps`, viewerToken.token
    );
    expect(appsResult.status).toBe(200);
    
    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];
      
      // Test read access to specific application details
      const appResult = await oidcHelpers.makeAuthenticatedRequest(
        request, 'GET', `${baseUrl}/api/apps/${firstApp.id}`, viewerToken.token
      );
      expect(appResult.status).toBe(200);
      expect(appResult.data).toBeTruthy();
    }
  });

  test('viewer should have read access to groups', async ({ request }) => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Get applications first
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps?page=0&perpage=10`, viewerToken.token
    );
    expect(appsResult.status).toBe(200);
    
    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];
      
      // Test read access to groups
      const groupsResult = await oidcHelpers.makeAuthenticatedRequest(
        request, 'GET', `${baseUrl}/api/apps/${firstApp.id}/groups`, viewerToken.token
      );
      expect(groupsResult.status).toBe(200);
    }
  });

  test('viewer should be denied group creation', async ({ request }) => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Get applications first
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps?page=0&perpage=10`, viewerToken.token
    );
    expect(appsResult.status).toBe(200);
    
    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];
      
      // try to create a new group
      const newGroup = {
        name: 'Viewer Test Group',
        policy_max_updates_per_period: 1,
        policy_period_interval: '15 minutes'
      };
      
      const createResult = await oidcHelpers.makeAuthenticatedRequest(
        request, 'POST', `${baseUrl}/api/apps/${firstApp.id}/groups`, viewerToken.token, newGroup
      );
      
      // viewer should be forbidden for write operations
      expect(createResult.status).toBe(403);
    }
  });

  test('viewer should have read access to packages', async ({ request }) => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Get applications first
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps?page=0&perpage=10`, viewerToken.token
    );
    expect(appsResult.status).toBe(200);
    
    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];
      
      // Test read access to packages
      const packagesResult = await oidcHelpers.makeAuthenticatedRequest(
        request, 'GET', `${baseUrl}/api/apps/${firstApp.id}/packages`, viewerToken.token
      );
      expect(packagesResult.status).toBe(200);
    }
  });

  test('viewer should be denied package creation', async ({ request }) => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Get applications first
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps?page=0&perpage=10`, viewerToken.token
    );
    expect(appsResult.status).toBe(200);
    
    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];
      
      // try to create a new package
      const newPackage = {
        application_id: firstApp.id,
        arch: 1,
        type: 1,
        url: 'https://example.com/package.gz',
        version: '1.0.0'
      };
      
      const createResult = await oidcHelpers.makeAuthenticatedRequest(
        request, 'POST', `${baseUrl}/api/apps/${firstApp.id}/packages`, viewerToken.token, newPackage
      );
      
      // viewer should be forbidden for write operations
      expect(createResult.status).toBe(403);
    }
  });

  test('viewer should have read access to instances', async ({ request }) => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Get applications first
    const appsResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps?page=0&perpage=10`, viewerToken.token
    );
    expect(appsResult.status).toBe(200);
    
    if (appsResult.data?.applications?.length > 0) {
      const firstApp = appsResult.data.applications[0];
      
      if (firstApp.groups?.length > 0) {
        const firstGroup = firstApp.groups[0];
        
        // viewer should have read access to instances
        const instancesResult = await oidcHelpers.makeAuthenticatedRequest(
          request, 'GET', `${baseUrl}/api/apps/${firstApp.id}/groups/${firstGroup.id}/instances`, viewerToken.token
        );
        expect(instancesResult.status).toBe(200);
      }
    }
  });

  test('viewer should have read access to activity logs', async ({ request }) => {
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // viewer should have read access to activity
    const activityResult = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/activity?start=${new Date(Date.now() - 24*60*60*1000).toISOString()}&end=${new Date().toISOString()}`, viewerToken.token
    );
    expect(activityResult.status).toBe(200);
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
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // compare admin vs viewer permissions
    const { adminResult, viewerResult } = await oidcHelpers.testRoleBasedAccess(
      request, '/api/apps', 'GET'
    );
    
    // both should be able to read
    expect(adminResult.status).toBe(200);
    expect(viewerResult.status).toBe(200);
    
    // test write operation permissions
    const newApp = { name: 'Role Test App' };
    
    const adminToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const viewerToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    
    const adminWrite = await oidcHelpers.makeAuthenticatedRequest(
      request, 'POST', `${baseUrl}/api/apps`, adminToken.token, newApp
    );
    const viewerWrite = await oidcHelpers.makeAuthenticatedRequest(
      request, 'POST', `${baseUrl}/api/apps`, viewerToken.token, newApp
    );
    
    // admin should be authorized
    expect(adminWrite.status).not.toBe(403);
    expect(adminWrite.status).not.toBe(401);
    
    // viewer should be forbidden
    expect(viewerWrite.status).toBe(403);
  });
});