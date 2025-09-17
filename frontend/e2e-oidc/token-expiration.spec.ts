import { expect, test } from '@playwright/test';

import { OIDCHelpers } from './helpers/oidc-helpers';
import { TEST_USERS } from './helpers/test-users';
import { OIDC_TEST_CONFIG } from './test-config';

test.describe('OIDC Token Expiration', () => {
  let oidcHelpers: OIDCHelpers;

  test.beforeEach(async () => {
    oidcHelpers = new OIDCHelpers();
  });

  test('should reject expired tokens', async ({ request }) => {
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // Get an expired token (mock)
    const expiredToken = await oidcHelpers.tokenManager.getExpiredToken(TEST_USERS.VIEWER);

    // Try to access protected API with expired token
    const result = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps`,
      expiredToken.token
    );

    // Should be unauthorized due to expiration
    expect(result.status).toBe(401);
  });

  test('should handle token expiration during request', async ({ request }) => {
    // This test simulates a scenario where a token expires between
    // the time it's obtained and when it's used

    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // Get a short-lived token (if available) or regular token
    const tokenInfo = await oidcHelpers.tokenManager.getShortLivedToken(TEST_USERS.VIEWER);

    // Verify token is initially valid
    expect(tokenInfo.isExpired).toBe(false);

    // Use token immediately (should work)
    const immediateResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps`,
      tokenInfo.token
    );
    expect(immediateResult.status).toBe(200);

    // For testing purposes, we'll use a mock expired token instead of waiting
    const expiredToken = await oidcHelpers.tokenManager.getExpiredToken(TEST_USERS.VIEWER);

    // Try to use expired token
    const expiredResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps`,
      expiredToken.token
    );
    expect(expiredResult.status).toBe(401);
  });

  test('should validate token expiration timestamps', async () => {
    const tokenInfo = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);

    // Verify token has expiration claim
    expect(tokenInfo.payload.exp).toBeTruthy();
    expect(typeof tokenInfo.payload.exp).toBe('number');

    // Verify expiration is in the future
    const now = Math.floor(Date.now() / 1000);
    expect(tokenInfo.payload.exp).toBeGreaterThan(now);

    // Verify issued at time is in the past or present
    expect(tokenInfo.payload.iat).toBeLessThanOrEqual(now);

    // Verify expiration date conversion
    expect(tokenInfo.expiresAt).toBeInstanceOf(Date);
    expect(tokenInfo.expiresAt.getTime()).toBeGreaterThan(Date.now());
  });

  test('should handle concurrent requests with expired tokens', async ({ request }) => {
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // Create multiple expired tokens
    const expiredTokens = await Promise.all([
      oidcHelpers.tokenManager.getExpiredToken(TEST_USERS.VIEWER),
      oidcHelpers.tokenManager.getExpiredToken(TEST_USERS.ADMIN),
      oidcHelpers.tokenManager.getExpiredToken(TEST_USERS.VIEWER),
    ]);

    // Make concurrent requests with expired tokens
    const requests = expiredTokens.map(tokenInfo =>
      oidcHelpers.makeAuthenticatedRequest(request, 'GET', `${baseUrl}/api/apps`, tokenInfo.token)
    );

    const results = await Promise.all(requests);

    // All should be unauthorized
    results.forEach(result => {
      expect(result.status).toBe(401);
    });
  });

  test('should compare valid vs expired token behavior', async ({ request }) => {
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // Get valid and expired tokens for the same user
    const validToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const expiredToken = await oidcHelpers.tokenManager.getExpiredToken(TEST_USERS.ADMIN);

    // Test with valid token
    const validResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps`,
      validToken.token
    );
    expect(validResult.status).toBe(200);
    expect(validResult.data).toBeTruthy();

    // Test with expired token
    const expiredResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps`,
      expiredToken.token
    );
    expect(expiredResult.status).toBe(401);
  });

  test('should handle different expired token formats', async ({ request }) => {
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // Test various forms of "expired" tokens
    const testTokens = [
      // Completely invalid token
      'expired.token.invalid',

      // JWT with obviously invalid expiration
      'eyJhbGciOiJSUzI1NiJ9.eyJleHAiOjB9.invalid',

      // Get mock expired token
      (await oidcHelpers.tokenManager.getExpiredToken(TEST_USERS.VIEWER)).token,
    ];

    for (const token of testTokens) {
      const result = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'GET',
        `${baseUrl}/api/apps`,
        token
      );

      // All should be unauthorized (401) for invalid/expired tokens
      expect(result.status).toBe(401);
    }
  });

  test('should handle token without expiration claim', async ({ request }) => {
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // Create a malformed JWT without exp claim
    const headerPayload = {
      alg: 'RS256',
      typ: 'JWT',
    };

    const payloadWithoutExp = {
      sub: '12345',
      name: 'Test User',
      // Missing 'exp' claim
      iat: Math.floor(Date.now() / 1000),
    };

    const header = btoa(JSON.stringify(headerPayload)).replace(/=/g, '');
    const payload = btoa(JSON.stringify(payloadWithoutExp)).replace(/=/g, '');
    const tokenWithoutExp = `${header}.${payload}.invalid-signature`;

    const result = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps`,
      tokenWithoutExp
    );

    // Should be rejected due to missing expiration or invalid signature
    expect(result.status).toBe(401);
  });

  test('should validate token refresh scenario simulation', async ({ request }) => {
    // This test simulates what would happen in a token refresh scenario
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // Simulate old expired token
    const expiredToken = await oidcHelpers.tokenManager.getExpiredToken(TEST_USERS.ADMIN);

    // Verify expired token is rejected
    const expiredResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps`,
      expiredToken.token
    );
    expect(expiredResult.status).toBe(401);

    // Simulate getting new fresh token (like after refresh)
    const freshToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);

    // Verify fresh token works
    const freshResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps`,
      freshToken.token
    );
    expect(freshResult.status).toBe(200);

    // Verify both tokens have same user but different expiration
    expect(expiredToken.payload?.preferred_username).toBe(freshToken.payload.preferred_username);
    if (expiredToken.payload?.exp && freshToken.payload.exp) {
      expect(freshToken.payload.exp).toBeGreaterThan(expiredToken.payload.exp);
    }
  });
});
