import { APIRequestContext,expect, Page } from '@playwright/test';

import { KeycloakAPI } from './keycloak-api';
import { TEST_USERS,TestUser } from './test-users';
import { TokenInfo,TokenManager } from './token-manager';

export class OIDCHelpers {
  private _tokenManager: TokenManager;
  private keycloakAPI: KeycloakAPI;

  constructor() {
    this._tokenManager = new TokenManager();
    this.keycloakAPI = new KeycloakAPI();
  }

  /**
   * Get the token manager instance
   */
  get tokenManager(): TokenManager {
    return this._tokenManager;
  }

  /**
   * Make an authenticated API request using a bearer token
   */
  async makeAuthenticatedRequest(
    request: APIRequestContext,
    method: 'GET' | 'POST' | 'PUT' | 'DELETE',
    url: string,
    token: string,
    data?: any
  ): Promise<any> {
    const response = await request.fetch(url, {
      method,
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      data: data ? JSON.stringify(data) : undefined,
    });

    return {
      status: response.status(),
      data: response.ok() ? await response.json().catch(() => null) : null,
      response
    };
  }

  /**
   * Test API endpoint protection with various token scenarios
   */
  async testAPIProtection(request: APIRequestContext, endpoint: string) {
    
    // Test without token
    const noTokenResponse = await request.get(endpoint);
    expect(noTokenResponse.status()).toBe(403); // Should be forbidden without token

    // Test with invalid token
    const invalidResult = await this.makeAuthenticatedRequest(
      request, 'GET', endpoint, 'invalid-token'
    );
    expect(invalidResult.status).toBe(401); // Should be unauthorized with invalid token

    // Test with malformed token
    const malformedResult = await this.makeAuthenticatedRequest(
      request, 'GET', endpoint, 'malformed.jwt.token'
    );
    expect(malformedResult.status).toBe(401); // Should be unauthorized with malformed token

    // Test with valid admin token
    const adminToken = await this._tokenManager.getValidToken(TEST_USERS.ADMIN);
    const adminResult = await this.makeAuthenticatedRequest(
      request, 'GET', endpoint, adminToken.token
    );
    expect(adminResult.status).toBe(200); // Should be OK with valid admin token

    return adminResult.data;
  }

  /**
   * Test role-based access control
   */
  async testRoleBasedAccess(request: APIRequestContext, endpoint: string, method: 'GET' | 'POST' | 'PUT' | 'DELETE' = 'GET') {
    
    // Test admin access
    const adminToken = await this.tokenManager.getValidToken(TEST_USERS.ADMIN);
    const adminResult = await this.makeAuthenticatedRequest(
      request, method, endpoint, adminToken.token
    );

    // Test viewer access
    const viewerToken = await this.tokenManager.getValidToken(TEST_USERS.VIEWER);
    const viewerResult = await this.makeAuthenticatedRequest(
      request, method, endpoint, viewerToken.token
    );

    return { adminResult, viewerResult };
  }

  /**
   * Simulate expired token scenario
   */
  async testExpiredToken(request: APIRequestContext, endpoint: string) {
    
    const expiredToken = await this.tokenManager.getExpiredToken(TEST_USERS.VIEWER);
    const result = await this.makeAuthenticatedRequest(
      request, 'GET', endpoint, expiredToken.token
    );
    
    expect(result.status).toBe(401); // Should be unauthorized with expired token
    return result;
  }

  /**
   * Verify token contains expected claims
   */
  verifyTokenClaims(tokenInfo: TokenInfo, expectedUser: TestUser) {
    expect(tokenInfo.payload).toBeTruthy();
    expect(tokenInfo.payload.preferred_username).toBe(expectedUser.username);
    
    // Check roles are present
    const roles = tokenInfo.payload.realm_access?.roles || [];
    for (const expectedRole of expectedUser.roles) {
      expect(roles).toContain(expectedRole);
    }
  }

  /**
   * Wait for UI to reflect authentication state
   */
  async waitForAuthenticationState(page: Page, isAuthenticated: boolean, timeout: number = 10000) {
    if (isAuthenticated) {
      // Wait for authenticated UI elements to appear
      await expect(page.locator('[data-testid="user-menu"], .user-info, .logout-button')).toBeVisible({ timeout });
    } else {
      // Wait for login form or unauthenticated state
      await expect(page.locator('[data-testid="login-form"], .login-button, .auth-required')).toBeVisible({ timeout });
    }
  }

  /**
   * Simulate authentication flow (for frontend tests)
   */
  async simulateAuthentication(page: Page, user: TestUser) {
    // This would be used to simulate the frontend authentication flow
    // In a real implementation, this might involve:
    // 1. Navigating to login page
    // 2. Redirecting to Keycloak
    // 3. Filling login form
    // 4. Handling redirect back to app
    // 5. Storing token in localStorage/sessionStorage
    
    const tokenInfo = await this.tokenManager.getValidToken(user);
    
    // For testing purposes, we can inject the token directly into the browser
    await page.addInitScript((token) => {
      localStorage.setItem('access_token', token);
    }, tokenInfo.token);
    
    return tokenInfo;
  }

  /**
   * Clear authentication state
   */
  async clearAuthentication(page: Page) {
    await page.evaluate(() => {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('id_token');
      sessionStorage.clear();
    });
  }

  /**
   * Get current authentication state from browser
   */
  async getAuthenticationState(page: Page): Promise<{ 
    token: string | null; 
    isAuthenticated: boolean; 
  }> {
    const token = await page.evaluate(() => localStorage.getItem('access_token'));
    return {
      token,
      isAuthenticated: !!token
    };
  }

  /**
   * Test OIDC discovery endpoints
   */
  async testOIDCDiscovery() {
    const discovery = await this.keycloakAPI.getOIDCDiscovery();
    expect(discovery).toBeTruthy();
    expect(discovery['token-service']).toContain('/protocol/openid-connect');
    
    const jwks = await this.keycloakAPI.getJWKS();
    expect(jwks.keys).toBeTruthy();
    expect(jwks.keys.length).toBeGreaterThan(0);
    
    return { discovery, jwks };
  }
}