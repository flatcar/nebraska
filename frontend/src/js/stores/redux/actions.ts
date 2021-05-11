import { SET_CONFIG } from "./actionTypes";

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

export function setConfig(config: NebraskaConfig): ConfigAction {
  return {
    type: SET_CONFIG,
    ...config
  };
};
