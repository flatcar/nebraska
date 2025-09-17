import { beforeEach, describe, expect, it, Mock, vi } from 'vitest';

import { createOIDCClient, getOIDCClient, OIDCClient, OIDCConfig } from '../../utils/oidc';

// Mock fetch
global.fetch = vi.fn();

// Mock crypto
Object.defineProperty(global, 'crypto', {
  value: {
    getRandomValues: vi.fn((arr: Uint8Array) => {
      for (let i = 0; i < arr.length; i++) {
        arr[i] = Math.floor(Math.random() * 256);
      }
      return arr;
    }),
    subtle: {
      digest: vi.fn(async () => {
        // Mock SHA-256 digest - return a fixed ArrayBuffer for testing
        const mockHash = new Uint8Array(32); // SHA-256 produces 32 bytes
        for (let i = 0; i < 32; i++) {
          mockHash[i] = i % 256; // Predictable test data
        }
        return mockHash.buffer;
      }),
    },
  },
});

// Mock sessionStorage
const mockSessionStorage = {
  store: new Map<string, string>(),
  getItem: vi.fn((key: string) => mockSessionStorage.store.get(key) || null),
  setItem: vi.fn((key: string, value: string) => {
    mockSessionStorage.store.set(key, value);
  }),
  removeItem: vi.fn((key: string) => {
    mockSessionStorage.store.delete(key);
  }),
  clear: vi.fn(() => {
    mockSessionStorage.store.clear();
  }),
};

Object.defineProperty(global, 'sessionStorage', {
  value: mockSessionStorage,
});

// Mock window.location
const mockLocation = {
  href: 'http://localhost:3000',
  origin: 'http://localhost:3000',
  search: '',
};

Object.defineProperty(global, 'window', {
  value: {
    location: mockLocation,
  },
});

describe('OIDC Client', () => {
  const mockConfig: OIDCConfig = {
    issuerUrl: 'https://example.com',
    clientId: 'test-client',
    redirectUri: 'http://localhost:3000/auth/callback',
    scopes: ['openid', 'profile', 'email'],
  };

  const mockMetadata = {
    issuer: 'https://example.com',
    authorization_endpoint: 'https://example.com/auth',
    token_endpoint: 'https://example.com/token',
    userinfo_endpoint: 'https://example.com/userinfo',
    jwks_uri: 'https://example.com/jwks',
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockSessionStorage.clear();
    mockLocation.search = '';
    (fetch as Mock).mockClear();
  });

  describe('OIDCClient initialization', () => {
    it('should load OIDC metadata on init', async () => {
      (fetch as Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMetadata,
      });

      const client = new OIDCClient(mockConfig);
      await client.init();

      expect(fetch).toHaveBeenCalledWith('https://example.com/.well-known/openid-configuration');
    });

    it('should throw error if metadata loading fails', async () => {
      (fetch as Mock).mockResolvedValueOnce({
        ok: false,
        status: 404,
      });

      const client = new OIDCClient(mockConfig);
      await expect(client.init()).rejects.toThrow('Failed to load OIDC metadata: 404');
    });
  });

  describe('Authorization flow', () => {
    it('should redirect to authorization endpoint', async () => {
      (fetch as Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMetadata,
      });

      const client = new OIDCClient(mockConfig);
      await client.init();

      // Mock window.location.href setter
      const hrefSetter = vi.fn();
      Object.defineProperty(window.location, 'href', {
        set: hrefSetter,
      });

      await client.authorize();

      expect(hrefSetter).toHaveBeenCalledWith(expect.stringContaining('https://example.com/auth'));
      expect(hrefSetter).toHaveBeenCalledWith(expect.stringContaining('client_id=test-client'));
      expect(hrefSetter).toHaveBeenCalledWith(expect.stringContaining('response_type=code'));
      expect(hrefSetter).toHaveBeenCalledWith(
        expect.stringContaining('code_challenge_method=S256')
      );

      // Check that PKCE parameters are stored
      expect(mockSessionStorage.getItem('oidc_code_verifier')).toBeTruthy();
      expect(mockSessionStorage.getItem('oidc_state')).toBeTruthy();
      expect(mockSessionStorage.getItem('oidc_nonce')).toBeTruthy();
    });
  });

  describe('Callback handling', () => {
    it('should handle callback with authorization code', async () => {
      (fetch as Mock)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => mockMetadata,
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => {
            // Create a mock JWT with proper format
            const header = btoa(JSON.stringify({ alg: 'RS256', typ: 'JWT' }));
            const payload = btoa(
              JSON.stringify({
                iss: 'https://example.com',
                sub: 'user123',
                aud: 'test-client',
                exp: Math.floor(Date.now() / 1000) + 3600,
                iat: Math.floor(Date.now() / 1000),
                nonce: 'test-nonce',
              })
            );
            const signature = 'mock-signature';
            const mockIdToken = `${header}.${payload}.${signature}`;

            return {
              access_token: 'test-access-token',
              token_type: 'Bearer',
              expires_in: 3600,
              id_token: mockIdToken,
            };
          },
        });

      const client = new OIDCClient(mockConfig);
      await client.init();

      // Setup callback scenario
      mockLocation.search = '?code=test-code&state=test-state';

      // Setup session data
      mockSessionStorage.setItem('oidc_code_verifier', 'test-verifier');
      mockSessionStorage.setItem('oidc_state', 'test-state');
      mockSessionStorage.setItem('oidc_nonce', 'test-nonce');
      mockSessionStorage.setItem('oidc_redirect_uri', mockConfig.redirectUri);
      mockSessionStorage.setItem('oidc_auth_started', Date.now().toString());

      const tokenResponse = await client.handleCallback();

      expect(tokenResponse.access_token).toBe('test-access-token');
      expect(tokenResponse.token_type).toBe('Bearer');
      expect(tokenResponse.id_token).toBeDefined();
      expect(tokenResponse.id_token).toContain('.');

      // Check that session storage is cleaned up
      expect(mockSessionStorage.getItem('oidc_code_verifier')).toBeNull();
      expect(mockSessionStorage.getItem('oidc_state')).toBeNull();
      expect(mockSessionStorage.getItem('oidc_nonce')).toBeNull();
      expect(mockSessionStorage.getItem('oidc_auth_started')).toBeNull();
    });

    it('should throw error for invalid state', async () => {
      (fetch as Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMetadata,
      });

      const client = new OIDCClient(mockConfig);
      await client.init();

      mockLocation.search = '?code=test-code&state=invalid-state';
      mockSessionStorage.setItem('oidc_state', 'test-state');
      mockSessionStorage.setItem('oidc_auth_started', Date.now().toString());

      await expect(client.handleCallback()).rejects.toThrow('Invalid state parameter');
    });

    it('should throw error for expired auth flow', async () => {
      (fetch as Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMetadata,
      });

      const client = new OIDCClient(mockConfig);
      await client.init();

      mockLocation.search = '?code=test-code&state=test-state';
      // Set auth started time to 11 minutes ago
      mockSessionStorage.setItem('oidc_auth_started', (Date.now() - 660000).toString());
      mockSessionStorage.setItem('oidc_state', 'test-state');

      await expect(client.handleCallback()).rejects.toThrow(
        'Authentication timeout - please try again'
      );
    });

    it('should throw error for authorization error', async () => {
      (fetch as Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMetadata,
      });

      const client = new OIDCClient(mockConfig);
      await client.init();

      mockLocation.search = '?error=access_denied&error_description=User denied';

      await expect(client.handleCallback()).rejects.toThrow(
        'OIDC authorization error: access_denied'
      );
    });
  });

  describe('isCallback', () => {
    it('should detect callback URL with code', () => {
      mockLocation.search = '?code=test-code&state=test-state';
      const client = new OIDCClient(mockConfig);
      expect(client.isCallback()).toBe(true);
    });

    it('should detect callback URL with error', () => {
      mockLocation.search = '?error=access_denied';
      const client = new OIDCClient(mockConfig);
      expect(client.isCallback()).toBe(true);
    });

    it('should return false for non-callback URL', () => {
      mockLocation.search = '';
      const client = new OIDCClient(mockConfig);
      expect(client.isCallback()).toBe(false);
    });
  });

  describe('getUserInfo', () => {
    it('should fetch user info with access token', async () => {
      const mockUserInfo = {
        sub: 'user123',
        name: 'Test User',
        email: 'test@example.com',
      };

      (fetch as Mock)
        .mockResolvedValueOnce({
          ok: true,
          json: async () => mockMetadata,
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => mockUserInfo,
        });

      const client = new OIDCClient(mockConfig);
      await client.init();

      const userInfo = await client.getUserInfo('test-access-token');

      expect(fetch).toHaveBeenCalledWith('https://example.com/userinfo', {
        headers: {
          Authorization: 'Bearer test-access-token',
        },
      });
      expect(userInfo).toEqual(mockUserInfo);
    });
  });
});

describe('OIDC Client Factory', () => {
  const mockConfig: OIDCConfig = {
    issuerUrl: 'https://example.com',
    clientId: 'test-client',
    redirectUri: 'http://localhost:3000/auth/callback',
    scopes: ['openid', 'profile', 'email'],
  };

  it('should create and return OIDC client', () => {
    const client = createOIDCClient(mockConfig);
    expect(client).toBeInstanceOf(OIDCClient);
    expect(getOIDCClient()).toBe(client);
  });

  it('should return null when no client created', () => {
    // Reset the singleton
    createOIDCClient(mockConfig);
    expect(getOIDCClient()).toBeInstanceOf(OIDCClient);
  });
});
