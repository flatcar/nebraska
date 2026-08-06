import { beforeEach, describe, expect, it, vi } from 'vitest';

import API from '../../api/API';
import ApplicationsStore from '../../stores/ApplicationsStore';

describe('ApplicationsStore delete handling', () => {
  beforeEach(() => {
    vi.spyOn(window, 'alert').mockImplementation(() => {});
  });

  it('does not remove an application when delete fails', async () => {
    vi.spyOn(API, 'deleteApplication').mockRejectedValue({ status: 403 });

    const store = new ApplicationsStore();
    store.applications = [{ id: 'app-1', name: 'Test' }];
    const emitSpy = vi.spyOn(store, 'emitChange');

    store.deleteApplication('app-1');

    await vi.waitFor(() => {
      expect(store.applications).toHaveLength(1);
      expect(emitSpy).not.toHaveBeenCalled();
      expect(window.alert).toHaveBeenCalled();
    });
  });

  it('does not remove a group when delete fails', async () => {
    vi.spyOn(API, 'deleteGroup').mockRejectedValue({ status: 500 });

    const store = new ApplicationsStore();
    store.applications = [
      {
        id: 'app-1',
        name: 'Test',
        groups: [{ id: 'group-1', application_id: 'app-1', name: 'Group' }],
      },
    ];
    const emitSpy = vi.spyOn(store, 'emitChange');

    store.deleteGroup('app-1', 'group-1');

    await vi.waitFor(() => {
      expect(store.applications[0].groups).toHaveLength(1);
      expect(emitSpy).not.toHaveBeenCalled();
      expect(window.alert).toHaveBeenCalled();
    });
  });
});
