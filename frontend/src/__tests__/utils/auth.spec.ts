import { jwtDecode as jwt_decode } from 'jwt-decode';
import { beforeEach, describe, expect, it, Mock, vi } from 'vitest';

import {
  clearToken,
  clearTokens,
  generateCodeChallenge,
  generateCodeVerifier,
  getToken,
  isValidToken,
  setToken,
  setTokens,
} from '../../utils/auth';

vi.mock('jwt-decode', () => ({
  jwtDecode: vi.fn(),
}));

describe('Auth Utility Functions', () => {
  beforeEach(() => {
    // Clear memory storage
    clearTokens();
    vi.clearAllMocks();
  });

  describe('setToken', () => {
    it('should store the token in memory', () => {
      const token = 'test-token';
      setToken(token);
      expect(getToken()).toBe(token);
    });
  });

  describe('getToken', () => {
    it('should retrieve the token from memory', () => {
      const token = 'test-token';
      setToken(token);
      expect(getToken()).toBe(token);
    });

    it('should return null if no token is stored', () => {
      expect(getToken()).toBeNull();
    });
  });

  describe('setTokens', () => {
    it('should store both access and refresh tokens', () => {
      const tokens = {
        access_token: 'access-token',
        expires_in: 3600,
      };
      setTokens(tokens);
      expect(getToken()).toBe('access-token');
    });

    it('should store access token without refresh token', () => {
      const tokens = {
        access_token: 'access-token',
        expires_in: 3600,
      };
      setTokens(tokens);
      expect(getToken()).toBe('access-token');
    });
  });

  describe('clearToken', () => {
    it('should clear the token from memory', () => {
      setToken('test-token');
      clearToken();
      expect(getToken()).toBeNull();
    });
  });

  describe('clearTokens', () => {
    it('should clear both access and refresh tokens', () => {
      setTokens({
        access_token: 'access-token',
        expires_in: 3600,
      });
      clearTokens();
      expect(getToken()).toBeNull();
    });
  });

  describe('isValidToken', () => {
    it('should return false for an empty token', () => {
      expect(isValidToken('')).toBe(false);
    });

    it('should return false for an expired token', () => {
      const expiredToken = 'expired-token';
      (jwt_decode as Mock).mockReturnValue({ exp: Math.floor(Date.now() / 1000) - 10 });
      expect(isValidToken(expiredToken)).toBe(false);
    });

    it('should return true for a valid token', () => {
      const validToken = 'valid-token';
      (jwt_decode as Mock).mockReturnValue({ exp: Math.floor(Date.now() / 1000) + 3600 });
      expect(isValidToken(validToken)).toBe(true);
    });
  });

  describe('PKCE functions', () => {
    describe('generateCodeVerifier', () => {
      it('should generate a code verifier of proper length', async () => {
        const verifier = await generateCodeVerifier();
        expect(verifier).toBeTruthy();
        expect(verifier.length).toBeGreaterThanOrEqual(43);
        expect(verifier.length).toBeLessThanOrEqual(128);
        // Should be URL-safe base64 (no +, /, or =)
        expect(verifier).not.toMatch(/[+/=]/);
      });
    });

    describe('generateCodeChallenge', () => {
      it('should generate a valid code challenge from verifier', async () => {
        const verifier = await generateCodeVerifier();
        const challenge = await generateCodeChallenge(verifier);
        expect(challenge).toBeTruthy();
        // Should be URL-safe base64 (no +, /, or =)
        expect(challenge).not.toMatch(/[+/=]/);
      });
    });
  });
});
