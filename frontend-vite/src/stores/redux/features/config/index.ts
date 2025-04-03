import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export const CONFIG_STORAGE_KEY = 'nebraska_config';

export interface NebraskaConfig {
  title: string;
  nebraska_config: string;
  nebraska_version: string;
  login_url: string;
  auth_mode: string;
  header_style: string;
  [prop: string]: any;
}

const nebraskaConfig = localStorage.getItem(CONFIG_STORAGE_KEY) || '{}';
const initialState: NebraskaConfig = JSON.parse(nebraskaConfig) || ({} as NebraskaConfig);

export const configSlice = createSlice({
  name: 'config',
  initialState,
  reducers: {
    setConfig: (state, action: PayloadAction<NebraskaConfig>) => {
      Object.assign(state, action.payload);
      localStorage.setItem(CONFIG_STORAGE_KEY, JSON.stringify(state));
    },
  },
});

export const { setConfig } = configSlice.actions;
export default configSlice.reducer;
