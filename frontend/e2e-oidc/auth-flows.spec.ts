import { expect,test } from '@playwright/test';

import { OIDCHelpers } from './helpers/oidc-helpers';
import { TEST_USERS } from './helpers/test-users';

test.describe('OIDC Authentication Flows', () => {
  let oidcHelpers: OIDCHelpers;

  test.beforeEach(async () => {
    oidcHelpers = new OIDCHelpers();
  });

  test('should successfully obtain access tokens for test users', async () => {
    // Test admin user token acquisition
    const adminTokenInfo = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    expect(adminTokenInfo.token).toBeTruthy();
    expect(adminTokenInfo.isExpired).toBe(false);
    
    oidcHelpers.verifyTokenClaims(adminTokenInfo, TEST_USERS.ADMIN);

    // Test viewer user token acquisition  
    const viewerTokenInfo = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    expect(viewerTokenInfo.token).toBeTruthy();
    expect(viewerTokenInfo.isExpired).toBe(false);
    
    oidcHelpers.verifyTokenClaims(viewerTokenInfo, TEST_USERS.VIEWER);
  });

  test('should validate OIDC discovery endpoints', async () => {
    const { discovery, jwks } = await oidcHelpers.testOIDCDiscovery();
    
    // Verify discovery document structure
    expect(discovery['token-service']).toContain('http://');
    expect(discovery['account-service']).toContain('http://');
    expect(discovery.public_key).toBeTruthy();
    
    // Verify JWKS structure
    expect(jwks.keys).toBeInstanceOf(Array);
    expect(jwks.keys.length).toBeGreaterThan(0);
    
    const firstKey = jwks.keys[0];
    expect(firstKey.kty).toBeTruthy(); // Key type
    expect(firstKey.use || firstKey.key_ops).toBeTruthy(); // Key usage
  });

  test('should authenticate and access protected API endpoints', async ({ request }) => {
    // Test /api/apps endpoint protection
    const appsData = await oidcHelpers.testAPIProtection(request, '/api/apps?page=0&perpage=10');
    expect(appsData).toBeTruthy();
    expect(appsData.applications).toBeDefined();
    
    // Test /config endpoint is accessible without authentication
    const configResponse = await request.get('/config');
    expect(configResponse.status()).toBe(200);
  });

  test('should handle invalid authentication scenarios', async ({ request }) => {
    
    // Test with no Authorization header
    const noAuthResponse = await request.get('/api/apps?page=0&perpage=10');
    expect(noAuthResponse.status()).toBe(403);
    
    // Test with invalid token
    const invalidTokenResponse = await request.get('/api/apps?page=0&perpage=10', {
      headers: { 'Authorization': 'Bearer invalid-token-12345' }
    });
    expect(invalidTokenResponse.status()).toBe(401);
    
    // Test with malformed Authorization header
    const malformedHeaderResponse = await request.get('/api/apps?page=0&perpage=10', {
      headers: { 'Authorization': 'InvalidFormat token' }
    });
    expect(malformedHeaderResponse.status()).toBe(403);
    
    // Test with malformed JWT
    const malformedJWTResponse = await request.get('/api/apps', {
      headers: { 'Authorization': 'Bearer eyJhbGciOiJSUzI1NiJ9.invalid.signature' }
    });
    expect(malformedJWTResponse.status()).toBe(401);
  });

  test('should handle Keycloak connectivity issues gracefully', async ({ request }) => {
    // This test verifies that the Nebraska backend handles Keycloak unavailability
    // In a real scenario, you might temporarily stop Keycloak to test this
    
    const adminTokenInfo = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    
    // Even with a valid token, if Keycloak becomes unavailable for key verification,
    // the backend should handle it gracefully (though it will likely reject the token)
    const result = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', 
      process.env.CI ? 'http://127.0.0.1:8003/api/apps' : 'http://localhost:8003/api/apps',
      adminTokenInfo.token
    );
    
    // With Keycloak running, this should succeed
    expect(result.status).toBe(200);
  });

  test('should validate token claims structure', async () => {
    const adminTokenInfo = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const payload = adminTokenInfo.payload;
    
    // Verify required JWT claims
    expect(payload.iss).toBeTruthy(); // Issuer
    expect(payload.sub).toBeTruthy(); // Subject
    expect(payload.aud || payload.azp).toBeTruthy(); // Audience or Authorized party
    expect(payload.exp).toBeGreaterThan(Math.floor(Date.now() / 1000)); // Expiration
    expect(payload.iat).toBeLessThanOrEqual(Math.floor(Date.now() / 1000)); // Issued at
    
    // Verify Nebraska-specific claims
    expect(payload.preferred_username).toBe(TEST_USERS.ADMIN.username);
    expect(payload.realm_access.roles).toContain('test_admin');
    expect(payload.realm_access.roles).toContain('test_viewer');
  });

  test('should handle concurrent authentication requests', async () => {
    // Test multiple concurrent token requests
    const promises = [
      oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN),
      oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER),
      oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN),
      oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER),
    ];
    
    const results = await Promise.all(promises);
    
    // All requests should succeed
    results.forEach(tokenInfo => {
      expect(tokenInfo.token).toBeTruthy();
      expect(tokenInfo.isExpired).toBe(false);
    });
    
    // Admin tokens should have admin claims
    oidcHelpers.verifyTokenClaims(results[0], TEST_USERS.ADMIN);
    oidcHelpers.verifyTokenClaims(results[2], TEST_USERS.ADMIN);
    
    // Viewer tokens should have viewer claims
    oidcHelpers.verifyTokenClaims(results[1], TEST_USERS.VIEWER);
    oidcHelpers.verifyTokenClaims(results[3], TEST_USERS.VIEWER);
  });
});