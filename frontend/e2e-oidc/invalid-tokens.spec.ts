import { expect, test } from '@playwright/test';

import { OIDCHelpers } from './helpers/oidc-helpers';
import { TEST_USERS } from './helpers/test-users';
import { OIDC_TEST_CONFIG } from './test-config';

test.describe('OIDC Invalid Token Handling', () => {
  let oidcHelpers: OIDCHelpers;

  test.beforeEach(async () => {
    oidcHelpers = new OIDCHelpers();
  });

  test('should reject invalid and malformed tokens', async ({ request }) => {
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    const invalidTokens = [
      // Completely invalid tokens
      'invalid-token',
      'completely.fake.token',
      '12345',
      'random-string-that-is-not-jwt',

      // Malformed JWT tokens
      'eyJhbGciOiJSUzI1NiJ9', // Missing parts
      'eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature.extra', // Too many parts
      'invalid-header.invalid-payload.invalid-signature', // Invalid base64

      // Empty string
      '',
    ];

    for (const invalidToken of invalidTokens) {
      const result = await oidcHelpers.makeAuthenticatedRequest(
        request,
        'GET',
        `${baseUrl}/api/apps?page=0&perpage=10`,
        invalidToken
      );

      // Should be unauthorized (401) for invalid tokens
      expect(result.status).toBe(401);
    }
  });

  test('should reject tokens with tampered content', async ({ request }) => {
    const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;

    // Get a valid token and tamper with its components
    const validToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const parts = validToken.token.split('.');

    // Test signature tampering
    const tamperedSignatureToken = `${parts[0]}.${parts[1]}.tampered-signature`;
    const signatureResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps`,
      tamperedSignatureToken
    );
    expect(signatureResult.status).toBe(401);

    // Test payload tampering (privilege escalation attempt)
    const tamperedPayload = {
      sub: 'hacker',
      preferred_username: 'admin',
      realm_access: { roles: ['test_admin', 'super_admin'] },
      exp: Math.floor(Date.now() / 1000) + 3600,
    };
    const tamperedPayloadEncoded = btoa(JSON.stringify(tamperedPayload)).replace(/=/g, '');
    const tamperedPayloadToken = `${parts[0]}.${tamperedPayloadEncoded}.${parts[2]}`;

    const payloadResult = await oidcHelpers.makeAuthenticatedRequest(
      request,
      'GET',
      `${baseUrl}/api/apps`,
      tamperedPayloadToken
    );
    expect(payloadResult.status).toBe(401);
  });
});
