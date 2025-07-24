import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export interface UserState {
  name?: string;
  email?: string;
  authenticated?: boolean;
  authLoading?: boolean;
  [prop: string]: any;
}

const initialState: UserState = {
  authLoading: true,
};

export const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    setUser: (state, action: PayloadAction<UserState>) => {
      Object.assign(state, action.payload);
    },
    setAuthLoading: (state, action: PayloadAction<boolean>) => {
      state.authLoading = action.payload;
    },
  },
});

export const { setUser, setAuthLoading } = userSlice.actions;
export default userSlice.reducer;
