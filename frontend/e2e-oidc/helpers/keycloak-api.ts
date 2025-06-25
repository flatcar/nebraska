import { OIDC_TEST_CONFIG } from '../test-config';
import { TestUser } from './test-users';

export class KeycloakAPI {
  private baseUrl: string;
  private realm: string;
  private clientId: string;

  constructor() {
    this.baseUrl = OIDC_TEST_CONFIG.keycloak.baseURL;
    this.realm = OIDC_TEST_CONFIG.keycloak.realm;
    this.clientId = OIDC_TEST_CONFIG.keycloak.clientId;
  }

  async getAccessToken(user: TestUser): Promise<string> {
    const tokenUrl = `${this.baseUrl}/realms/${this.realm}/protocol/openid-connect/token`;

    const response = await fetch(tokenUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: new URLSearchParams({
        client_id: this.clientId,
        username: user.username,
        password: user.password,
        grant_type: 'password',
        scope: 'openid profile email',
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to get access token: ${response.status} ${errorText}`);
    }

    const data = await response.json();
    return data.access_token;
  }

  async getTokenWithCustomExpiry(user: TestUser): Promise<string> {
    // For testing token expiration, we'll get a regular token
    // In a real scenario, you might use a custom token endpoint or mock
    const token = await this.getAccessToken(user);

    // For now, return the regular token
    // In practice, you would need to either:
    // 1. Configure Keycloak with very short token lifetimes
    // 2. Use a custom token generation service for tests
    // 3. Mock the token validation in the backend for tests
    return token;
  }

  async getOIDCDiscovery(): Promise<any> {
    const discoveryUrl = `${this.baseUrl}/realms/${this.realm}`;
    const response = await fetch(discoveryUrl);

    if (!response.ok) {
      throw new Error(`Failed to get OIDC discovery: ${response.status}`);
    }

    return response.json();
  }

  async getJWKS(): Promise<any> {
    const jwksUrl = `${this.baseUrl}/realms/${this.realm}/protocol/openid-connect/certs`;
    const response = await fetch(jwksUrl);

    if (!response.ok) {
      throw new Error(`Failed to get JWKS: ${response.status}`);
    }

    return response.json();
  }

  /**
   * Create an invalid token for testing purposes
   */
  createInvalidToken(): string {
    return 'invalid.token.signature';
  }

  /**
   * Create a malformed JWT token for testing
   */
  createMalformedToken(): string {
    return 'eyJhbGciOiJSUzI1NiJ9.invalid.signature';
  }

  /**
   * Decode JWT payload without verification (for testing purposes only)
   */
  decodeJWTPayload(token: string): any {
    const parts = token.split('.');
    if (parts.length !== 3) {
      throw new Error('Invalid JWT format');
    }

    const payload = parts[1];
    const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
    return JSON.parse(decoded);
  }
}
