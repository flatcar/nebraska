import { configureStore } from '@reduxjs/toolkit';

import configReducer from './features/config';
import userReducer from './features/user';

export function makeStore() {
  return configureStore({
    reducer: {
      config: configReducer,
      user: userReducer,
    },
  });
}

const store = makeStore();

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

export default store;
