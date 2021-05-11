import { TypedUseSelectorHook, useSelector } from 'react-redux';
import { ConfigAction, NebraskaConfig } from '../actions';
import { SET_CONFIG } from '../actionTypes';

const CONFIG_STORAGE_KEY = 'nebraska_config';

export interface State {
  config: NebraskaConfig;
}

const initialState: State = {
  config: JSON.parse(localStorage.getItem(CONFIG_STORAGE_KEY) || "") as NebraskaConfig,
};

type Action = ConfigAction |
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
    default:
      return state;
  }
}

export type RootState = ReturnType<typeof reducer>;

export const useTypedSelector: TypedUseSelectorHook<RootState> = useSelector;

export default reducer;
