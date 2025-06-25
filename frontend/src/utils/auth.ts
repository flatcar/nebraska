import { jwtDecode } from 'jwt-decode';
import React from 'react';
import { useLocation, useNavigate } from 'react-router';

import { setUser, UserState } from '../stores/redux/features/user';
import { useDispatch, useSelector } from '../stores/redux/hooks';
import { createOIDCClient, OIDCConfig } from './oidc';

// In-memory token storage for better security
let accessToken: string | null = null;
let idToken: string | null = null;

interface TokenPair {
  access_token: string;
  expires_in?: number;
  id_token?: string;
}

export function setTokens(tokens: TokenPair) {
  accessToken = tokens.access_token;
  if (tokens.id_token) {
    idToken = tokens.id_token;
  }
}

export function getToken() {
  return accessToken;
}

export function clearTokens() {
  accessToken = null;
  idToken = null;
}

export function getIdToken() {
  return idToken;
}

// Legacy function for backward compatibility
export function setToken(token: string) {
  accessToken = token;
}

export function clearToken() {
  clearTokens();
}

// Check if we have a valid token
export function ensureValidToken(): string | null {
  if (!accessToken) {
    return null;
  }

  // If token is still valid, return it
  if (isValidToken(accessToken)) {
    return accessToken;
  }

  // Token is expired, clear it
  clearTokens();
  return null;
}

// PKCE helper functions
export async function generateCodeVerifier(): Promise<string> {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return base64URLEncode(array);
}

export async function generateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await crypto.subtle.digest('SHA-256', data);
  return base64URLEncode(new Uint8Array(digest));
}

function base64URLEncode(buffer: Uint8Array): string {
  const base64 = btoa(String.fromCharCode(...buffer));
  return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

interface JWT {
  exp: number;
  [prop: string]: any;
}

export function isValidToken(token: string) {
  if (token === '') {
    return false;
  }

  const decoded = jwtDecode(token) as JWT;

  // Check if it's expired
  const expiration = new Date(decoded.exp * 1000);
  if (expiration < new Date()) {
    return false;
  }

  return true;
}

function getUserInfoFromToken(token: string) {
  const info: UserState = {
    name: '',
    email: '',
  };

  if (token === '') {
    return info;
  }

  const decoded = jwtDecode(token) as JWT;

  // Try multiple claims for name in order of preference
  info.name = decoded.name || decoded.given_name || decoded.preferred_username || '';
  info.email = decoded.email || '';

  return info;
}

export function useAuthRedirect() {
  const config = useSelector(state => state.config);
  const user = useSelector(state => state.user);
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const location = useLocation();

  const shouldUpdateUser = React.useCallback(
    (token: string) => {
      const newInfo = getUserInfoFromToken(token);

      for (const [key, value] of Object.entries(newInfo)) {
        if (user[key] !== value) {
          return true;
        }
      }

      return false;
    },
    [user]
  );

  React.useEffect(() => {
    // We only do the login dance if the auth mode is OIDC
    if (config.auth_mode !== 'oidc') {
      return;
    }

    const initOIDC = async () => {
      // Create OIDC configuration from config
      const configuredScopes = config.oidc_scopes || 'openid,profile,email';
      const oidcConfig: OIDCConfig = {
        issuerUrl: config.oidc_issuer_url || '',
        clientId: config.oidc_client_id || '',
        redirectUri: window.location.origin + '/auth/callback',
        scopes: configuredScopes.split(',').map((s: string) => s.trim()),
        logoutUrl: config.oidc_logout_url,
        audience: config.oidc_audience,
      };

      // Create OIDC client
      const oidcClient = createOIDCClient(oidcConfig);
      await oidcClient.init();

      // Check if this is a callback from OIDC provider
      if (oidcClient.isCallback()) {
        try {
          const tokenResponse = await oidcClient.handleCallback();

          // Store access token and ID token
          setTokens({
            access_token: tokenResponse.access_token,
            expires_in: tokenResponse.expires_in,
            id_token: tokenResponse.id_token,
          });

          // Get user info from userinfo endpoint
          const userInfo: UserState = { authenticated: true };
          try {
            const userInfoResponse = await oidcClient.getUserInfo(tokenResponse.access_token);
            userInfo.name = userInfoResponse.name || userInfoResponse.given_name || '';
            userInfo.email = userInfoResponse.email || '';
          } catch (error) {
            console.warn('Failed to get user info:', error);
          }

          dispatch(setUser(userInfo));

          // Clean up URL and redirect to the original location
          navigate(tokenResponse.returnUrl || '/', { replace: true });
          return;
        } catch (error) {
          console.error('OIDC callback error:', error);
          // Clear stored tokens and redirect to login
          clearTokens();
          dispatch(setUser({ authenticated: false }));

          // Redirect to root to avoid being stuck on callback URL
          navigate('/', { replace: true });
        }
      }

      // Check if user is already authenticated
      const currentToken = ensureValidToken();

      if (currentToken && shouldUpdateUser(currentToken)) {
        dispatch(setUser({ authenticated: true, ...getUserInfoFromToken(currentToken) }));
        return;
      } else if (currentToken) {
        // Token is valid but user info hasn't changed
        return;
      }

      // If not authenticated and we have OIDC config, start authorization
      if (!user?.authenticated && oidcConfig.issuerUrl && oidcConfig.clientId) {
        await oidcClient.authorize();
      }
    };

    initOIDC().catch(console.error);
  }, [navigate, location, user, config, dispatch, shouldUpdateUser]);
}
