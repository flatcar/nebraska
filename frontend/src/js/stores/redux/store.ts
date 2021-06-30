import { configureStore } from '@reduxjs/toolkit';
import configReducer from './features/config';
import userReducer from './features/user';

const store = configureStore({
  reducer: {
    config: configReducer,
    user: userReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

export default store;
