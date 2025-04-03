import { jwtDecode as jwt_decode } from 'jwt-decode';
import { beforeEach, describe, expect, it, Mock, vi } from 'vitest';

import { getToken, isValidToken, setToken } from '../../utils/auth';

vi.mock('jwt-decode', () => ({
  jwtDecode: vi.fn(),
}));

describe('Auth Utility Functions', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  describe('setToken', () => {
    it('should store the token in localStorage', () => {
      const token = 'test-token';
      setToken(token);
      expect(localStorage.getItem('token')).toBe(token);
    });
  });

  describe('getToken', () => {
    it('should retrieve the token from localStorage', () => {
      const token = 'test-token';
      localStorage.setItem('token', token);
      expect(getToken()).toBe(token);
    });

    it('should return null if no token is stored', () => {
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
});
