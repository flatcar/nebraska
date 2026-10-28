import { TFunction } from 'i18next';

import { useSelector } from '../stores/redux/hooks';

export type InstanceMode = 'single' | 'control' | 'edge';

export function instanceModeOf(config: { instance_mode?: string }): InstanceMode {
  const mode = config.instance_mode;
  return mode === 'control' || mode === 'edge' ? mode : 'single';
}

export function useInstanceMode(): InstanceMode {
  return useSelector(state => instanceModeOf(state.config));
}

export function adminWriteRefusal(t: TFunction, mode: InstanceMode): string | undefined {
  return mode === 'edge' ? t('common|admin_writes_unavailable') : undefined;
}

export function hasLocalOverrides(mode: InstanceMode): mode is Exclude<InstanceMode, 'single'> {
  return mode !== 'single';
}
