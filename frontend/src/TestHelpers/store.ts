import { NebraskaConfig, setConfig } from '../stores/redux/features/config';
import { makeStore as makeAppStore } from '../stores/redux/store';

// A store like the app's, holding the /config of a node of the given mode, or
// of an older server that reports no mode.
export function makeStore(instanceMode?: string) {
  const store = makeAppStore();
  if (instanceMode) {
    store.dispatch(setConfig({ instance_mode: instanceMode } as NebraskaConfig));
  }
  return store;
}
