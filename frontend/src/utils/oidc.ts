import { generateCodeChallenge, generateCodeVerifier } from './auth';

export interface OIDCConfig {
  issuerUrl: string;
  clientId: string;
  redirectUri: string;
  scopes: string[];
  logoutUrl?: string;
}

export interface TokenResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  id_token?: string;
  scope?: string;
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

  constructor(config: OIDCConfig) {
    this.config = config;
  }

  async init(): Promise<void> {
    await this.loadMetadata();
  }

  private async loadMetadata(): Promise<void> {
    const wellKnownUrl = `${this.config.issuerUrl}/.well-known/openid_configuration`;
    
    try {
      const response = await fetch(wellKnownUrl);
      if (!response.ok) {
        throw new Error(`Failed to load OIDC metadata: ${response.status}`);
      }
      this.metadata = await response.json();
    } catch (error) {
      console.error('Failed to load OIDC metadata:', error);
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

    // Store PKCE verifier and state for callback validation
    sessionStorage.setItem('oidc_code_verifier', codeVerifier);
    sessionStorage.setItem('oidc_state', state);
    sessionStorage.setItem('oidc_redirect_uri', this.config.redirectUri);

    const scopes = [...this.config.scopes];

    const params = new URLSearchParams({
      client_id: this.config.clientId,
      response_type: 'code',
      redirect_uri: this.config.redirectUri,
      scope: scopes.join(' '),
      code_challenge: codeChallenge,
      code_challenge_method: 'S256',
      state: state,
    });

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

    if (error) {
      throw new Error(`OIDC authorization error: ${error}`);
    }

    if (!code) {
      throw new Error('Authorization code not found in callback');
    }

    // Validate state
    const storedState = sessionStorage.getItem('oidc_state');
    if (!storedState || state !== storedState) {
      throw new Error('Invalid state parameter');
    }

    // Get stored PKCE verifier
    const codeVerifier = sessionStorage.getItem('oidc_code_verifier');
    const redirectUri = sessionStorage.getItem('oidc_redirect_uri');
    
    if (!codeVerifier || !redirectUri) {
      throw new Error('PKCE verifier or redirect URI not found');
    }

    // Exchange code for tokens
    const tokenResponse = await this.exchangeCodeForTokens(code, codeVerifier, redirectUri);

    // Clean up session storage
    sessionStorage.removeItem('oidc_code_verifier');
    sessionStorage.removeItem('oidc_state');
    sessionStorage.removeItem('oidc_redirect_uri');

    return tokenResponse;
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


  async logout(postLogoutRedirectUri?: string): Promise<void> {
    if (!this.metadata) {
      await this.init();
    }

    if (this.metadata.end_session_endpoint) {
      const params = new URLSearchParams();
      if (postLogoutRedirectUri) {
        params.set('post_logout_redirect_uri', postLogoutRedirectUri);
      }
      
      const logoutUrl = `${this.metadata.end_session_endpoint}?${params}`;
      window.location.href = logoutUrl;
    } else if (postLogoutRedirectUri) {
      // If no end_session_endpoint, just redirect to post logout URI
      window.location.href = postLogoutRedirectUri;
    }
  }

  private generateState(): string {
    const array = new Uint8Array(16);
    crypto.getRandomValues(array);
    return btoa(String.fromCharCode(...array))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
  }

  isCallback(): boolean {
    const params = new URLSearchParams(window.location.search);
    return params.has('code') || params.has('error');
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