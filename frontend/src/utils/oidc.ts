import {
  base64ToBase64URL,
  base64URLEncode,
  generateCodeChallenge,
  generateCodeVerifier,
} from './auth';

export interface OIDCConfig {
  issuerUrl: string;
  clientId: string;
  redirectUri: string;
  scopes: string[];
  logoutUrl?: string;
  audience?: string;
}

export interface TokenResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  id_token?: string;
  scope?: string;
  returnUrl?: string;
}

export interface OIDCProviderMetadata {
  issuer: string;
  authorization_endpoint: string;
  token_endpoint: string;
  userinfo_endpoint: string;
  end_session_endpoint?: string;
  jwks_uri: string;
}

export class OIDCClient {
  private config: OIDCConfig;
  private metadata: OIDCProviderMetadata | null = null;
  private metadataLoadedAt: number = 0;
  private readonly METADATA_TTL_MS = 3600000; // 1 hour

  constructor(config: OIDCConfig) {
    this.config = config;
  }

  async init(): Promise<void> {
    await this.loadMetadata();
  }

  private async loadMetadata(): Promise<void> {
    // Check if metadata is still fresh
    if (this.metadata && Date.now() - this.metadataLoadedAt < this.METADATA_TTL_MS) {
      return;
    }

    const wellKnownUrl = `${this.config.issuerUrl.replace(/\/$/, '')}/.well-known/openid-configuration`;

    const errMsg = (err: any) => `Failed to load OIDC metadata: ${err}`;
    try {
      const response = await fetch(wellKnownUrl);
      if (!response.ok) {
        throw new Error(errMsg(response.status));
      }
      this.metadata = await response.json();
      this.metadataLoadedAt = Date.now();
    } catch (error) {
      console.error(errMsg(error));
      throw error;
    }
  }

  async authorize(): Promise<void> {
    if (!this.metadata) {
      await this.init();
    }

    const codeVerifier = await generateCodeVerifier();
    const codeChallenge = await generateCodeChallenge(codeVerifier);
    const state = this.generateState();
    const nonce = this.generateNonce();

    // Store PKCE verifier, state, and nonce for callback validation
    sessionStorage.setItem('oidc_code_verifier', codeVerifier);
    sessionStorage.setItem('oidc_state', state);
    sessionStorage.setItem('oidc_nonce', nonce);
    sessionStorage.setItem('oidc_redirect_uri', this.config.redirectUri);

    // Store timestamp for cleanup of abandoned auth flows
    sessionStorage.setItem('oidc_auth_started', Date.now().toString());

    const scopes = [...this.config.scopes];

    const params = new URLSearchParams({
      client_id: this.config.clientId,
      response_type: 'code',
      redirect_uri: this.config.redirectUri,
      scope: scopes.join(' '),
      code_challenge: codeChallenge,
      code_challenge_method: 'S256',
      state: state,
      nonce: nonce,
    });

    if (this.config.audience) {
      params.set('audience', this.config.audience);
    }

    const authUrl = `${this.metadata!.authorization_endpoint}?${params}`;
    window.location.href = authUrl;
  }

  async handleCallback(): Promise<TokenResponse> {
    if (!this.metadata) {
      await this.init();
    }

    const params = new URLSearchParams(window.location.search);
    const code = params.get('code');
    const state = params.get('state');
    const error = params.get('error');
    const errorDescription = params.get('error_description');

    if (error) {
      throw new Error(
        `OIDC authorization error: ${error}${errorDescription ? ` - ${errorDescription}` : ''}`
      );
    }

    if (!code) {
      throw new Error('Authorization code not found in callback');
    }

    // Check if auth flow is too old (>10 minutes)
    const authStarted = sessionStorage.getItem('oidc_auth_started');
    if (authStarted && Date.now() - parseInt(authStarted) > 600000) {
      // Clean up stale auth data
      this.clearAuthSession();
      throw new Error('Authentication timeout - please try again');
    }

    // Validate state
    const storedState = sessionStorage.getItem('oidc_state');
    if (!storedState || !state) {
      throw new Error('Invalid state parameter');
    }

    // Decode and validate state
    let returnUrl = '/';
    try {
      const decodedState = JSON.parse(atob(state.replace(/-/g, '+').replace(/_/g, '/')));
      const decodedStoredState = JSON.parse(
        atob(storedState.replace(/-/g, '+').replace(/_/g, '/'))
      );

      if (decodedState.nonce !== decodedStoredState.nonce) {
        throw new Error('State nonce mismatch');
      }

      returnUrl = decodedState.returnUrl || '/';
    } catch {
      // Fallback to simple string comparison for backward compatibility
      if (state !== storedState) {
        throw new Error('Invalid state parameter');
      }
    }

    // Get stored PKCE verifier
    const codeVerifier = sessionStorage.getItem('oidc_code_verifier');
    const redirectUri = sessionStorage.getItem('oidc_redirect_uri');

    if (!codeVerifier || !redirectUri) {
      throw new Error('PKCE verifier or redirect URI not found');
    }

    // Exchange code for tokens
    const tokenResponse = await this.exchangeCodeForTokens(code, codeVerifier, redirectUri);

    // Validate ID token if present
    if (tokenResponse.id_token) {
      const nonce = sessionStorage.getItem('oidc_nonce');
      this.validateIdToken(tokenResponse.id_token, nonce);
    }

    // Clean up session storage
    this.clearAuthSession();

    // Add the return URL to the token response
    return { ...tokenResponse, returnUrl };
  }

  private async exchangeCodeForTokens(
    code: string,
    codeVerifier: string,
    redirectUri: string
  ): Promise<TokenResponse> {
    const tokenEndpoint = this.metadata!.token_endpoint;

    const body = new URLSearchParams({
      grant_type: 'authorization_code',
      code: code,
      redirect_uri: redirectUri,
      client_id: this.config.clientId,
      code_verifier: codeVerifier,
    });

    const response = await fetch(tokenEndpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: body,
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Token exchange failed: ${response.status} ${errorText}`);
    }

    const tokenResponse: TokenResponse = await response.json();

    // Validate required fields in token response
    if (!tokenResponse.access_token || !tokenResponse.token_type) {
      throw new Error('Invalid token response: missing required fields');
    }

    return tokenResponse;
  }

  async getUserInfo(accessToken: string): Promise<any> {
    if (!this.metadata) {
      await this.init();
    }

    const response = await fetch(this.metadata!.userinfo_endpoint, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    });

    if (!response.ok) {
      throw new Error(`UserInfo request failed: ${response.status}`);
    }

    return response.json();
  }

  async logout(postLogoutRedirectUri?: string, idTokenHint?: string): Promise<void> {
    if (!this.metadata) {
      await this.init();
    }

    const logoutEndpoint = this.metadata?.end_session_endpoint || this.config.logoutUrl;

    if (logoutEndpoint) {
      const params = new URLSearchParams();
      if (postLogoutRedirectUri) {
        params.set('post_logout_redirect_uri', postLogoutRedirectUri);
      }
      if (idTokenHint) {
        params.set('id_token_hint', idTokenHint);
      }

      const logoutUrl = `${logoutEndpoint}?${params}`;
      window.location.href = logoutUrl;
    } else if (postLogoutRedirectUri) {
      // If no logout endpoint available, just redirect to post logout URI
      window.location.href = postLogoutRedirectUri;
    }
  }

  private generateState(): string {
    const array = new Uint8Array(16);
    crypto.getRandomValues(array);
    const randomValue = base64URLEncode(array);

    // Encode the current URL and random value in the state
    const stateData = {
      nonce: randomValue,
      returnUrl: window.location.pathname + window.location.search + window.location.hash,
    };

    return base64ToBase64URL(btoa(JSON.stringify(stateData)));
  }

  private generateNonce(): string {
    const array = new Uint8Array(16);
    crypto.getRandomValues(array);
    return base64URLEncode(array);
  }

  isCallback(): boolean {
    const params = new URLSearchParams(window.location.search);
    return params.has('code') || params.has('error');
  }

  private clearAuthSession(): void {
    sessionStorage.removeItem('oidc_code_verifier');
    sessionStorage.removeItem('oidc_state');
    sessionStorage.removeItem('oidc_nonce');
    sessionStorage.removeItem('oidc_redirect_uri');
    sessionStorage.removeItem('oidc_auth_started');
  }

  private validateIdToken(idToken: string, expectedNonce: string | null): void {
    // Basic JWT structure validation
    const parts = idToken.split('.');
    if (parts.length !== 3) {
      throw new Error('Invalid ID token format');
    }

    // Decode and validate claims
    // NOTE: This validates claims only. Signature validation is handled by the backend
    // which has access to the JWKS endpoint and can properly verify the cryptographic signature
    try {
      const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')));

      // Validate issuer
      if (payload.iss !== this.config.issuerUrl) {
        throw new Error(`Invalid issuer: ${payload.iss}`);
      }

      // Validate audience
      const audience = Array.isArray(payload.aud) ? payload.aud : [payload.aud];
      if (!audience.includes(this.config.clientId)) {
        throw new Error(`Invalid audience: ${payload.aud}`);
      }

      // Validate nonce
      if (expectedNonce && payload.nonce !== expectedNonce) {
        throw new Error('Invalid nonce in ID token');
      }

      // Validate expiration
      if (payload.exp && Date.now() >= payload.exp * 1000) {
        throw new Error('ID token expired');
      }

      // Validate issued at (not in future with 1 minute clock skew)
      if (payload.iat && payload.iat * 1000 > Date.now() + 60000) {
        throw new Error('ID token issued in the future');
      }
    } catch (error) {
      if (error instanceof Error) {
        throw error;
      }
      throw new Error('Failed to decode ID token');
    }
  }
}

// Create singleton instance
let oidcClient: OIDCClient | null = null;

export function createOIDCClient(config: OIDCConfig): OIDCClient {
  oidcClient = new OIDCClient(config);
  return oidcClient;
}

export function getOIDCClient(): OIDCClient | null {
  return oidcClient;
}
