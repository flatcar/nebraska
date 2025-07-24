import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export interface NebraskaConfig {
  title: string;
  nebraska_config: string;
  nebraska_version: string;
  login_url: string;
  auth_mode: string;
  header_style: string;
  [prop: string]: any;
}

const initialState: NebraskaConfig = {} as NebraskaConfig;

export const configSlice = createSlice({
  name: 'config',
  initialState,
  reducers: {
    setConfig: (state, action: PayloadAction<NebraskaConfig>) => {
      Object.assign(state, action.payload);
    },
  },
});

export const { setConfig } = configSlice.actions;
export default configSlice.reducer;
