import jwt_decode from "jwt-decode";
import React from "react";
import { useHistory } from "react-router";
import { useTypedSelector } from "../stores/redux/reducers";

const TOKEN_KEY = 'token';

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

interface JWT {
  exp: number;
  [prop: string]: any;
}

export function isValidToken(token: string) {
  if (token === '') {
    return false;
  }

  const decoded = jwt_decode(token) as JWT;

  // Check if it's expired
  const expiration = new Date(decoded.exp * 1000);
  if (expiration < (new Date())) {
    return false
  }

  return true;
}

export function useAuthRedirect() {
  const config = useTypedSelector(state => state.config);
  const history = useHistory();

  React.useEffect(() => {
    const params = new URLSearchParams(history.location.search);

    // We only do the login dance if the auth mode is OIDC
    if (config.auth_mode !== 'oidc') {
      return;
    }

    const token = params.get('id_token');
    if (!!token) {
      setToken(token);
      // Discard the URL search params
      history.push(history.location.pathname)
      return;
    }

    if (!isValidToken(getToken() || '') && !!config.login_url) {
      window.location.href = config.login_url + '?login_redirect_url=' + window.location.href;
    }
  },
  [history, config]);
}
