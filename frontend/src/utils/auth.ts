import { jwtDecode } from 'jwt-decode';
import { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router';

import { setAuthLoading, setUser, UserState } from '../stores/redux/features/user';
import { useDispatch, useSelector } from '../stores/redux/hooks';
import { authBroadcast } from './authBroadcast';
import { createOIDCClient, OIDCConfig } from './oidc';

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

// Export for testing purposes only
export function clearTokens() {
  accessToken = null;
  idToken = null;
}

export function broadcastLogout() {
  clearTokens();
  authBroadcast.broadcastLogout();
}

export function getIdToken() {
  return idToken;
}

export function ensureValidToken(): string | null {
  if (!accessToken) {
    return null;
  }

  if (isValidToken(accessToken)) {
    return accessToken;
  }

  broadcastLogout();
  return null;
}

export function ensureValidIdToken(): string | null {
  if (!idToken) {
    return null;
  }

  if (isValidToken(idToken)) {
    return idToken;
  }

  // If ID token is invalid but access token is still valid,
  // just clear the ID token, don't logout completely
  idToken = null;
  return null;
}

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

export function base64ToBase64URL(base64: string): string {
  return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

export function base64URLEncode(buffer: Uint8Array): string {
  const base64 = btoa(String.fromCharCode(...buffer));
  return base64ToBase64URL(base64);
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

  info.name = decoded.name || decoded.given_name || decoded.preferred_username || '';
  info.email = decoded.email || '';

  return info;
}

async function extractUserInfo(tokenResponse: TokenPair, oidcClient: any): Promise<UserState> {
  let userInfo: UserState = { authenticated: true };

  // First try to get info from ID token
  if (tokenResponse.id_token) {
    const idTokenInfo = getUserInfoFromToken(tokenResponse.id_token);
    userInfo = { ...userInfo, ...idTokenInfo };
  }

  // Only call userinfo endpoint if we're missing data
  if (!userInfo.name || !userInfo.email) {
    try {
      const userInfoResponse = await oidcClient.getUserInfo(tokenResponse.access_token);

      if (!userInfo.name && userInfoResponse.name) {
        userInfo.name = userInfoResponse.name;
      }

      if (!userInfo.email && userInfoResponse.email) {
        userInfo.email = userInfoResponse.email;
      }
    } catch (error) {
      console.warn('Failed to get user info:', error);
    }
  }

  return userInfo;
}

export function useAuthRedirect() {
  const config = useSelector(state => state.config);
  const user = useSelector(state => state.user);
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (config.auth_mode !== 'oidc') {
      // For non-OIDC modes (noop, github), no client-side auth flow needed
      dispatch(setAuthLoading(false));
      return;
    }

    const initOIDC = async () => {
      const configuredScopes = config.oidc_scopes || 'openid,profile,email';
      const parsedScopes = configuredScopes.split(/[,\s]+/).filter((s: string) => s.trim());

      const oidcConfig: OIDCConfig = {
        issuerUrl: config.oidc_issuer_url || '',
        clientId: config.oidc_client_id || '',
        redirectUri: window.location.origin + '/auth/callback',
        scopes: parsedScopes,
        logoutUrl: config.oidc_logout_url,
        audience: config.oidc_audience,
      };

      const oidcClient = createOIDCClient(oidcConfig);
      await oidcClient.init();

      if (oidcClient.isCallback()) {
        try {
          const tokenResponse = await oidcClient.handleCallback();
          setTokens({
            access_token: tokenResponse.access_token,
            expires_in: tokenResponse.expires_in,
            id_token: tokenResponse.id_token,
          });

          const userInfo = await extractUserInfo(tokenResponse, oidcClient);
          dispatch(setUser(userInfo));
          dispatch(setAuthLoading(false));

          navigate(tokenResponse.returnUrl || '/', { replace: true });
          return;
        } catch (error) {
          console.error('OIDC callback error:', error);
          broadcastLogout();
          dispatch(setUser({ authenticated: false }));
          dispatch(setAuthLoading(false));

          const errorMessage = error instanceof Error ? error.message : 'Authentication failed';
          navigate('/auth/error', {
            replace: true,
            state: { error: errorMessage },
          });
          return;
        }
      }

      if (ensureValidToken()) {
        dispatch(setAuthLoading(false));
        return;
      }

      if (!user?.authenticated && oidcConfig.issuerUrl && oidcConfig.clientId) {
        await oidcClient.authorize();
      }

      // Set loading to false at the end - we've either handled auth or will redirect
      dispatch(setAuthLoading(false));
    };

    initOIDC().catch(console.error);
  }, [navigate, location, user, config, dispatch]);
}

export function useLogoutSync() {
  const dispatch = useDispatch();

  useEffect(() => {
    return authBroadcast.onLogout(() => {
      clearTokens();

      // Clear user info on logout
      dispatch(setUser({ authenticated: false, name: '', email: '' }));
    });
  }, [dispatch]);
}
