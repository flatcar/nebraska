import { jwtDecode } from 'jwt-decode';
import { useCallback, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router';

import { setUser, UserState } from '../stores/redux/features/user';
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

export function useAuthRedirect() {
  const config = useSelector(state => state.config);
  const user = useSelector(state => state.user);
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const location = useLocation();

  const shouldUpdateUser = useCallback(
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

  useEffect(() => {
    if (config.auth_mode !== 'oidc') {
      return;
    }

    const initOIDC = async () => {
      const configuredScopes = config.oidc_scopes || 'openid,profile,email';
      const oidcConfig: OIDCConfig = {
        issuerUrl: config.oidc_issuer_url || '',
        clientId: config.oidc_client_id || '',
        redirectUri: window.location.origin + '/auth/callback',
        scopes: configuredScopes.split(',').map((s: string) => s.trim()),
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

          const userInfo: UserState = { authenticated: true };
          try {
            const userInfoResponse = await oidcClient.getUserInfo(tokenResponse.access_token);
            userInfo.name = userInfoResponse.name || userInfoResponse.given_name || '';
            userInfo.email = userInfoResponse.email || '';
          } catch (error) {
            console.warn('Failed to get user info:', error);
          }

          dispatch(setUser(userInfo));

          navigate(tokenResponse.returnUrl || '/', { replace: true });
          return;
        } catch (error) {
          console.error('OIDC callback error:', error);
          broadcastLogout();
          dispatch(setUser({ authenticated: false }));

          const errorMessage = error instanceof Error ? error.message : 'Authentication failed';
          navigate('/auth/error', {
            replace: true,
            state: { error: errorMessage },
          });
          return;
        }
      }

      const currentToken = ensureValidToken();

      if (currentToken && shouldUpdateUser(currentToken)) {
        dispatch(setUser({ authenticated: true, ...getUserInfoFromToken(currentToken) }));
        return;
      } else if (currentToken) {
        return;
      }

      if (!user?.authenticated && oidcConfig.issuerUrl && oidcConfig.clientId) {
        await oidcClient.authorize();
      }
    };

    initOIDC().catch(console.error);
  }, [navigate, location, user, config, dispatch, shouldUpdateUser]);
}

export function useAuthBroadcastSync() {
  const dispatch = useDispatch();

  useEffect(() => {
    return authBroadcast.onLogout(() => {
      clearTokens();

      dispatch(setUser({ authenticated: false, name: '', email: '' }));
    });
  }, [dispatch]);
}
