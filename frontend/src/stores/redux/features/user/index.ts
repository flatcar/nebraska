import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export interface UserState {
  name?: string;
  email?: string;
  authenticated?: boolean;
  [prop: string]: any;
}

const initialState: UserState = {};

export const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    setUser: (state, action: PayloadAction<UserState>) => {
      Object.assign(state, action.payload);
    },
  },
});

export const { setUser } = userSlice.actions;
export default userSlice.reducer;
