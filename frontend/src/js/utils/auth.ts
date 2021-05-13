import jwt_decode from "jwt-decode";
import React from "react";
import { useDispatch } from "react-redux";
import { useHistory } from "react-router";
import { setUser, UserState } from "../stores/redux/actions";
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

function getUserInfoFromToken(token: string) {
  const info: UserState = {
    name: '',
    email: '',
  };

  if (token === '') {
    return info;
  }

  const decoded = jwt_decode(token) as JWT;

  info.name = decoded.given_name || '';
  info.email = decoded.email || '';

  return info;
}

export function useAuthRedirect() {
  const config = useTypedSelector(state => state.config);
  const user = useTypedSelector(state => state.user);
  const dispatch = useDispatch();
  const history = useHistory();

  function shouldUpdateUser(token: string) {
    const newInfo = getUserInfoFromToken(token);

    for (const [key, value] of Object.entries(newInfo)) {
      if (user[key] !== value) {
        return true;
      }
    }

    return false;
  }

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
      dispatch(setUser({authenticated: true}));
      history.push(history.location.pathname)
      return;
    }

    const currentToken = getToken() || '';

    if (isValidToken(currentToken) && shouldUpdateUser(currentToken)) {
      dispatch(setUser({authenticated: true, ...getUserInfoFromToken(currentToken)}));
    }

    if ((!isValidToken(currentToken) || !user?.authenticated) && !!config.login_url) {
      window.location.href = config.login_url + '?login_redirect_url=' + window.location.href;
    }
  },
  [history, user, config]);
}
