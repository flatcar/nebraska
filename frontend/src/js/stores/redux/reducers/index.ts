import { TypedUseSelectorHook, useSelector } from 'react-redux';
import { ConfigAction, NebraskaConfig, SetUserAction, UserState } from '../actions';
import { SET_CONFIG, SET_USER } from '../actionTypes';

const CONFIG_STORAGE_KEY = 'nebraska_config';

export interface State {
  config: NebraskaConfig;
  user: UserState;
}

const initialState: State = {
  config: JSON.parse(localStorage.getItem(CONFIG_STORAGE_KEY) || "") as NebraskaConfig,
  user: {}
};

type Action =
  ConfigAction |
  SetUserAction |
{
  type: string;
  [prop: string]: any;
};

function reducer(state = initialState, action: Action) {
  const {type, ...actionProps} = action;
  switch (type) {
    case SET_CONFIG: {
      const newState = {...state};
      newState.config = {
        ...actionProps as NebraskaConfig,
      }
      localStorage.setItem(CONFIG_STORAGE_KEY, JSON.stringify(newState.config));
      return newState;
    }
    case SET_USER: {
      const newState = {...state};
      newState.user = {
        ...actionProps as UserState,
      }
      return newState;
    }
    default:
      return state;
  }
}

export type RootState = ReturnType<typeof reducer>;

export const useTypedSelector: TypedUseSelectorHook<RootState> = useSelector;

export default reducer;
