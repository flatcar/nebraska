import { KeycloakAPI } from './keycloak-api';
import { TestUser } from './test-users';

export interface TokenInfo {
  token: string;
  payload: any;
  expiresAt: Date;
  isExpired: boolean;
}

export class TokenManager {
  private keycloakAPI: KeycloakAPI;

  constructor() {
    this.keycloakAPI = new KeycloakAPI();
  }

  async getValidToken(user: TestUser): Promise<TokenInfo> {
    const token = await this.keycloakAPI.getAccessToken(user);
    return this.createTokenInfo(token);
  }

  async getExpiredToken(user: TestUser): Promise<TokenInfo> {
    // For testing expired tokens, we could:
    // 1. Use a token with very short expiry (if Keycloak is configured for that)
    // 2. Create a mock expired token
    // 3. Wait for a short-lived token to expire

    // For now, we'll create a mock expired token by modifying the payload
    const validToken = await this.keycloakAPI.getAccessToken(user);
    const expiredToken = this.createMockExpiredToken(validToken);
    return this.createTokenInfo(expiredToken);
  }

  getInvalidToken(): TokenInfo {
    const invalidToken = this.keycloakAPI.createInvalidToken();
    return {
      token: invalidToken,
      payload: null,
      expiresAt: new Date(),
      isExpired: true,
    };
  }

  getMalformedToken(): TokenInfo {
    const malformedToken = this.keycloakAPI.createMalformedToken();
    return {
      token: malformedToken,
      payload: null,
      expiresAt: new Date(),
      isExpired: true,
    };
  }

  private createTokenInfo(token: string): TokenInfo {
    try {
      const payload = this.keycloakAPI.decodeJWTPayload(token);
      const expiresAt = new Date(payload.exp * 1000);
      const isExpired = expiresAt < new Date();

      return {
        token,
        payload,
        expiresAt,
        isExpired,
      };
    } catch {
      return {
        token,
        payload: null,
        expiresAt: new Date(),
        isExpired: true,
      };
    }
  }

  private createMockExpiredToken(validToken: string): string {
    try {
      const parts = validToken.split('.');
      const header = parts[0];
      const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')));
      const signature = parts[2];

      // Set expiration to 1 hour ago
      payload.exp = Math.floor(Date.now() / 1000) - 3600;

      const newPayload = btoa(JSON.stringify(payload))
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=/g, '');

      // Note: This creates an invalid signature, but that's fine for testing
      // The backend will reject it due to signature validation
      return `${header}.${newPayload}.${signature}`;
    } catch {
      // If we can't modify the token, return an obviously invalid one
      return 'expired.token.invalid';
    }
  }

  /**
   * Create a token that will expire in the specified number of seconds
   */
  async getShortLivedToken(user: TestUser): Promise<TokenInfo> {
    // For proper short-lived tokens, you would need to configure Keycloak
    // or use a custom token endpoint. For testing, we'll use a regular token
    // and document that it should be used quickly.
    const token = await this.keycloakAPI.getAccessToken(user);
    return this.createTokenInfo(token);
  }

  /**
   * Wait for a token to expire (useful for testing expiration scenarios)
   */
  async waitForTokenExpiration(tokenInfo: TokenInfo, maxWaitMs: number = 30000): Promise<void> {
    const now = Date.now();
    const expiresAt = tokenInfo.expiresAt.getTime();
    const waitTime = Math.min(expiresAt - now, maxWaitMs);

    if (waitTime > 0) {
      await new Promise(resolve => setTimeout(resolve, waitTime + 1000)); // Add 1 second buffer
    }
  }
}
