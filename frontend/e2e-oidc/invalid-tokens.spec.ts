import { test, expect } from '@playwright/test';
import { OIDCHelpers } from './helpers/oidc-helpers';
import { TEST_USERS } from './helpers/test-users';

test.describe('OIDC Invalid Token Handling', () => {
  let oidcHelpers: OIDCHelpers;

  test.beforeEach(async () => {
    oidcHelpers = new OIDCHelpers();
  });

  test('should reject completely invalid tokens', async ({ request }) => {
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    const invalidTokens = [
      'invalid-token',
      'completely.fake.token',
      '12345',
      'Bearer token without Bearer prefix',
      'random-string-that-is-not-jwt',
    ];
    
    for (const invalidToken of invalidTokens) {
      const result = await oidcHelpers.makeAuthenticatedRequest(
        request, 'GET', `${baseUrl}/api/apps?page=0&perpage=10`, invalidToken
      );
      
      // Should be unauthorized (401) or forbidden (403) for invalid tokens
      expect([401, 403]).toContain(result.status);
    }
  });

  test('should reject malformed JWT tokens', async ({ request }) => {
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    const malformedTokens = [
      // Missing parts
      'eyJhbGciOiJSUzI1NiJ9',
      'eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0',
      
      // Too many parts
      'eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature.extra',
      
      // Invalid base64
      'invalid-header.invalid-payload.invalid-signature',
      
      // Valid structure but invalid JSON
      'aW52YWxpZC1qc29u.aW52YWxpZC1qc29u.invalid-signature',
    ];
    
    for (const malformedToken of malformedTokens) {
      const result = await oidcHelpers.makeAuthenticatedRequest(
        request, 'GET', `${baseUrl}/api/apps`, malformedToken
      );
      
      // Should be unauthorized due to malformed structure
      expect(result.status).toBe(401);
    }
  });

  test('should reject tokens with invalid signatures', async ({ request }) => {
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Get a valid token and tamper with its signature
    const validToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const parts = validToken.token.split('.');
    
    const tamperedTokens = [
      // Change signature
      `${parts[0]}.${parts[1]}.tampered-signature`,
      
      // Empty signature
      `${parts[0]}.${parts[1]}.`,
      
      // Wrong signature format
      `${parts[0]}.${parts[1]}.abc123`,
    ];
    
    for (const tamperedToken of tamperedTokens) {
      const result = await oidcHelpers.makeAuthenticatedRequest(
        request, 'GET', `${baseUrl}/api/apps`, tamperedToken
      );
      
      // Should be unauthorized due to signature verification failure
      expect(result.status).toBe(401);
    }
  });

  test('should reject tokens with tampered payload', async ({ request }) => {
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Get a valid token and tamper with its payload
    const validToken = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const parts = validToken.token.split('.');
    
    // Create tampered payload
    const tamperedPayload = {
      sub: 'hacker',
      preferred_username: 'admin',
      realm_access: {
        roles: ['test_admin', 'super_admin']
      },
      exp: Math.floor(Date.now() / 1000) + 3600 // 1 hour from now
    };
    
    const tamperedPayloadEncoded = btoa(JSON.stringify(tamperedPayload))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
    
    const tamperedToken = `${parts[0]}.${tamperedPayloadEncoded}.${parts[2]}`;
    
    const result = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps`, tamperedToken
    );
    
    // Should be unauthorized due to signature verification failure
    expect(result.status).toBe(401);
  });

  test('should reject tokens with wrong issuer', async ({ request }) => {
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Create a token with wrong issuer
    const wrongIssuerPayload = {
      iss: 'https://evil-server.com/realms/master',
      sub: 'user123',
      preferred_username: 'test-viewer',
      realm_access: {
        roles: ['test_viewer']
      },
      exp: Math.floor(Date.now() / 1000) + 3600,
      iat: Math.floor(Date.now() / 1000)
    };
    
    const header = btoa(JSON.stringify({ alg: 'RS256', typ: 'JWT' })).replace(/=/g, '');
    const payload = btoa(JSON.stringify(wrongIssuerPayload)).replace(/=/g, '');
    const wrongIssuerToken = `${header}.${payload}.fake-signature`;
    
    const result = await oidcHelpers.makeAuthenticatedRequest(
      request, 'GET', `${baseUrl}/api/apps`, wrongIssuerToken
    );
    
    // Should be unauthorized due to wrong issuer
    expect(result.status).toBe(401);
  });

  test('should reject tokens with missing required claims', async ({ request }) => {
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    const incompletePayloads = [
      // Missing subject
      {
        iss: 'http://keycloak:8080/realms/test',
        exp: Math.floor(Date.now() / 1000) + 3600,
        iat: Math.floor(Date.now() / 1000)
      },
      
      // Missing expiration
      {
        iss: 'http://keycloak:8080/realms/test',
        sub: 'user123',
        iat: Math.floor(Date.now() / 1000)
      },
      
      // Missing issued at
      {
        iss: 'http://keycloak:8080/realms/test',
        sub: 'user123',
        exp: Math.floor(Date.now() / 1000) + 3600
      }
    ];
    
    for (const payload of incompletePayloads) {
      const header = btoa(JSON.stringify({ alg: 'RS256', typ: 'JWT' })).replace(/=/g, '');
      const payloadEncoded = btoa(JSON.stringify(payload)).replace(/=/g, '');
      const incompleteToken = `${header}.${payloadEncoded}.fake-signature`;
      
      const result = await oidcHelpers.makeAuthenticatedRequest(
        request, 'GET', `${baseUrl}/api/apps`, incompleteToken
      );
      
      // Should be unauthorized due to missing required claims
      expect(result.status).toBe(401);
    }
  });

  test('should handle tokens with wrong algorithm', async ({ request }) => {
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Create token with wrong algorithm in header
    const wrongAlgHeaders = [
      { alg: 'HS256', typ: 'JWT' }, // Wrong algorithm (HMAC instead of RSA)
      { alg: 'none', typ: 'JWT' },  // No algorithm
      { typ: 'JWT' },               // Missing algorithm
    ];
    
    const payload = {
      iss: 'http://keycloak:8080/realms/test',
      sub: 'user123',
      preferred_username: 'test-viewer',
      exp: Math.floor(Date.now() / 1000) + 3600,
      iat: Math.floor(Date.now() / 1000)
    };
    
    for (const header of wrongAlgHeaders) {
      const headerEncoded = btoa(JSON.stringify(header)).replace(/=/g, '');
      const payloadEncoded = btoa(JSON.stringify(payload)).replace(/=/g, '');
      const wrongAlgToken = `${headerEncoded}.${payloadEncoded}.fake-signature`;
      
      const result = await oidcHelpers.makeAuthenticatedRequest(
        request, 'GET', `${baseUrl}/api/apps`, wrongAlgToken
      );
      
      // Should be unauthorized due to wrong algorithm
      expect(result.status).toBe(401);
    }
  });

  test('should handle empty and null tokens', async ({ request }) => {
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Test with empty token
    const emptyTokenResult = await request.get(`${baseUrl}/api/apps`, {
      headers: { 'Authorization': 'Bearer ' }
    });
    expect([401, 403]).toContain(emptyTokenResult.status());
    
    // Test with just "Bearer" 
    const bearerOnlyResult = await request.get(`${baseUrl}/api/apps`, {
      headers: { 'Authorization': 'Bearer' }
    });
    expect([401, 403]).toContain(bearerOnlyResult.status());
  });

  test('should handle authorization header edge cases', async ({ request }) => {
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    const edgeCaseHeaders = [
      'bearer invalid-token',           // Lowercase bearer
      'BEARER invalid-token',           // Uppercase bearer
      'Basic invalid-token',            // Wrong auth type
      'Bearer\tinvalid-token',          // Tab instead of space
      'Bearer  invalid-token',          // Double space
      'Bearer invalid-token extra',     // Extra content
    ];
    
    for (const authHeader of edgeCaseHeaders) {
      const result = await request.get(`${baseUrl}/api/apps`, {
        headers: { 'Authorization': authHeader }
      });
      
      // Should be unauthorized or forbidden
      expect([401, 403]).toContain(result.status());
    }
  });

  test('should handle concurrent invalid token requests', async ({ request }) => {
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Create multiple invalid tokens
    const invalidTokens = [
      'invalid-1',
      'invalid-2', 
      'malformed.jwt.token',
      oidcHelpers.tokenManager.getInvalidToken().token,
      oidcHelpers.tokenManager.getMalformedToken().token,
    ];
    
    // Make concurrent requests with invalid tokens
    const requests = invalidTokens.map(token =>
      oidcHelpers.makeAuthenticatedRequest(
        request, 'GET', `${baseUrl}/api/apps`, token
      )
    );
    
    const results = await Promise.all(requests);
    
    // All should be unauthorized
    results.forEach(result => {
      expect(result.status).toBe(401);
    });
  });

  test('should properly validate token structure before processing', async ({ request }) => {
    const baseUrl = process.env.CI ? 'http://127.0.0.1:8003' : 'http://localhost:8003';
    
    // Test various malformed JWT structures
    const structurallyInvalidTokens = [
      '',                    // Empty
      '.',                   // Just dots
      '..',                  // Just dots
      'a.',                  // Incomplete
      '.b.',                 // Incomplete
      'a.b.',               // Incomplete
      'a..c',               // Empty middle section
    ];
    
    for (const invalidToken of structurallyInvalidTokens) {
      const result = await oidcHelpers.makeAuthenticatedRequest(
        request, 'GET', `${baseUrl}/api/apps?page=0&perpage=10`, invalidToken
      );
      
      // Should be rejected early due to structural issues (401) or forbidden (403)
      expect([401, 403]).toContain(result.status);
    }
  });
});