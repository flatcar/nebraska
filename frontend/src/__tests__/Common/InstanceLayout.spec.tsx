import { act, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useNavigate } from 'react-router';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { Application, Instance } from '../../api/apiDataTypes';

const testState = vi.hoisted(() => ({
  getInstance: vi.fn(),
  addChangeListener: vi.fn(),
  removeChangeListener: vi.fn(),
  getApplication: vi.fn(),
  listener: undefined as (() => void) | undefined,
}));

vi.mock('../../api/API', () => ({
  default: {
    getInstance: testState.getInstance,
  },
}));

vi.mock('../../stores/Stores', () => ({
  applicationsStore: () => ({
    getCachedApplication: () => application,
    getCachedApplications: () => [application],
    addChangeListener: testState.addChangeListener,
    removeChangeListener: testState.removeChangeListener,
    getApplication: testState.getApplication,
  }),
}));

vi.mock('../../components/common/SectionHeader', () => ({
  default: () => null,
}));

vi.mock('../../components/common/Loader', () => ({
  default: () => <div data-testid="loader" />,
}));

vi.mock('../../components/Instances/Details', () => ({
  default: ({ instance }: { instance: Instance }) => (
    <div data-testid="instance-details">{instance.id}</div>
  ),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import InstanceLayout from '../../components/layouts/InstanceLayout/InstanceLayout';

const application = {
  id: 'app-1',
  name: 'Application',
  groups: [{ id: 'group-1', name: 'Group' }],
} as Application;

function makeInstance(id: string): Instance {
  return {
    id,
    alias: id,
    application: {
      status: null,
      version: '1.0.0',
    },
  } as Instance;
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>(resolvePromise => {
    resolve = resolvePromise;
  });
  return { promise, resolve };
}

function renderInstanceRoute() {
  function NavigateToInstanceB() {
    const navigate = useNavigate();
    return (
      <button onClick={() => navigate('/apps/app-1/groups/group-1/instances/instance-b')}>
        Show instance B
      </button>
    );
  }

  return render(
    <MemoryRouter initialEntries={['/apps/app-1/groups/group-1/instances/instance-a']}>
      <NavigateToInstanceB />
      <Routes>
        <Route
          path="/apps/:appID/groups/:groupID/instances/:instanceID"
          element={<InstanceLayout />}
        />
      </Routes>
    </MemoryRouter>
  );
}

describe('InstanceLayout', () => {
  beforeEach(() => {
    testState.getInstance.mockReset();
    testState.listener = undefined;
    testState.addChangeListener.mockReset();
    testState.removeChangeListener.mockReset();
    testState.getApplication.mockReset();
    testState.addChangeListener.mockImplementation(listener => {
      testState.listener = listener;
    });
    testState.getApplication.mockImplementation(() => {
      testState.listener?.();
    });
  });

  it('does not keep showing the previous instance while a new route loads', async () => {
    const instanceA = deferred<Instance>();
    const instanceB = deferred<Instance>();
    testState.getInstance
      .mockReturnValueOnce(instanceA.promise)
      .mockReturnValueOnce(instanceB.promise);

    renderInstanceRoute();

    await act(async () => instanceA.resolve(makeInstance('instance-a')));
    expect(screen.getByTestId('instance-details').textContent).toBe('instance-a');

    fireEvent.click(screen.getByText('Show instance B'));

    expect(screen.queryByTestId('instance-details')).toBeNull();
    expect(screen.getByTestId('loader')).toBeTruthy();
  });

  it('ignores a response from the previous instance route that arrives late', async () => {
    const instanceA = deferred<Instance>();
    const instanceB = deferred<Instance>();
    testState.getInstance
      .mockReturnValueOnce(instanceA.promise)
      .mockReturnValueOnce(instanceB.promise);

    renderInstanceRoute();

    fireEvent.click(screen.getByText('Show instance B'));

    await act(async () => instanceB.resolve(makeInstance('instance-b')));
    expect(screen.getByTestId('instance-details').textContent).toBe('instance-b');

    await act(async () => instanceA.resolve(makeInstance('instance-a')));
    expect(screen.getByTestId('instance-details').textContent).toBe('instance-b');
  });
});
