import { beforeEach, describe, expect, it, vi } from 'vitest';

import API from '../../api/API';
import { Application } from '../../api/apiDataTypes';
import ApplicationsStore from '../../stores/ApplicationsStore';
import { makeGroup } from '../../TestHelpers/groups';

describe('group local overrides API', () => {
  const mockFetch = vi.spyOn(global, 'fetch');

  beforeEach(() => {
    mockFetch.mockReset();
  });

  it('patches only the overrides it is given', async () => {
    mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ id: 'g' }) } as Response);

    const group = await API.updateGroupLocalOverrides('app', 'g', {
      policy_updates_enabled: null,
      policy_safe_mode: true,
    });

    expect(group).toEqual({ id: 'g' });
    const [url, init] = mockFetch.mock.calls[0];
    expect(url).toBe('/api/apps/app/groups/g/local_overrides');
    expect(init?.method).toBe('PATCH');
    expect(JSON.parse(init?.body as string)).toEqual({
      policy_updates_enabled: null,
      policy_safe_mode: true,
    });
  });

  it('hands a refused change back to the caller', async () => {
    const refused = { ok: false, status: 400 } as Response;
    mockFetch.mockResolvedValueOnce(refused);

    await expect(
      API.updateGroupLocalOverrides('app', 'g', { policy_safe_mode: true })
    ).rejects.toBe(refused);
  });
});

describe('ApplicationsStore.setGroupUpdatesPaused', () => {
  const mockUpdate = vi.spyOn(API, 'updateGroupLocalOverrides');

  beforeEach(() => {
    mockUpdate.mockReset();
  });

  it.each([
    [true, false],
    [false, null],
  ])(
    'paused=%s sends policy_updates_enabled=%s and keeps the group it gets back',
    async (paused, sent) => {
      const store = new ApplicationsStore();
      store.applications = [{ id: 'app-1', groups: [makeGroup()] } as unknown as Application];
      const updated = makeGroup({ policy_updates_enabled: paused ? false : null });
      mockUpdate.mockResolvedValueOnce(updated);
      const listener = vi.fn();
      store.addChangeListener(listener);

      await store.setGroupUpdatesPaused('app-1', 'group-1', paused);

      expect(mockUpdate).toHaveBeenCalledWith('app-1', 'group-1', { policy_updates_enabled: sent });
      expect(store.getCachedApplication('app-1')?.groups[0]).toBe(updated);
      expect(listener).toHaveBeenCalled();
    }
  );

  it('keeps the cached group when the node refuses', async () => {
    const store = new ApplicationsStore();
    const group = makeGroup();
    store.applications = [{ id: 'app-1', groups: [group] } as unknown as Application];
    const refused = { ok: false, status: 501 } as Response;
    mockUpdate.mockRejectedValueOnce(refused);

    await expect(store.setGroupUpdatesPaused('app-1', 'group-1', true)).rejects.toBe(refused);
    expect(store.getCachedApplication('app-1')?.groups[0]).toBe(group);
  });
});
