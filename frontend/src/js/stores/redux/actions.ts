import { SET_CONFIG, SET_USER } from "./actionTypes";

export interface NebraskaConfig {
  title: string;
  nebraska_config: string;
  nebraska_version: string;
  login_url: string;
  auth_mode: string;
  header_style: string;
  [prop: string]: any;
}

export interface ConfigAction extends NebraskaConfig {
  type: 'SET_CONFIG';
}

export interface UserState {
  name?: string;
  email?: string;
  authenticated?: boolean;
  [prop: string]: any;
}

export interface SetUserAction extends UserState {
  type: 'SET_USER';
}

export function setConfig(config: NebraskaConfig): ConfigAction {
  return {
    type: SET_CONFIG,
    ...config
  };
};

export function setUser(userState: UserState): SetUserAction {
  return {
    type: SET_USER,
    ...userState
  };
};
