
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
