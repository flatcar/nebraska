import '../../../i18n/config.ts';

import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { act, render, screen, waitFor } from '@testing-library/react';
import { ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { Group } from '../../../api/apiDataTypes';
import themes from '../../../lib/themes';
import { groupChartStoreContext } from '../../../stores/Stores';
import StatusCountTimeline from './StatusCountTimeline';
import { Duration } from './TimelineChart';
import VersionCountTimeline from './VersionCountTimeline';

vi.mock('./TimelineChart', () => ({
  default: (props: { data: any[]; keys: string[] }) => (
    <div data-testid="timeline-chart">
      {props.data.map(entry => (
        <div key={entry.timestamp}>
          {entry.timestamp}
          {props.keys.map(key => `${key}:${entry[key]}`).join(',')}
        </div>
      ))}
    </div>
  ),
}));

const duration: Duration = {
  displayValue: '1 day',
  queryValue: '1d',
  disabled: false,
};

function deferredTimeline<T>() {
  let resolve: (value: T) => void = () => {};
  const promise = new Promise<T>(res => {
    resolve = res;
  });

  return { promise, resolve };
}

function makeGroup(id: string, name: string): Group {
  return {
    id,
    name,
    description: '',
    created_ts: '2026-01-01T00:00:00Z',
    rollout_in_progress: false,
    application_id: 'app-1',
    channel_id: null,
    policy_updates_enabled: false,
    policy_safe_mode: false,
    policy_office_hours: false,
    policy_timezone: null,
    policy_period_interval: '',
    policy_max_updates_per_period: 0,
    policy_update_timeout: '',
    channel: {
      id: `channel-${id}`,
      name: 'stable',
      color: '#000000',
      created_ts: '2026-01-01T00:00:00Z',
      application_id: 'app-1',
      package_id: null,
      package: null,
      arch: 1,
    },
    track: 'stable',
  };
}

function wrapWithStore(ui: ReactNode, store: object) {
  const ChartStoreContext = groupChartStoreContext();

  return (
    <StyledEngineProvider injectFirst>
      <ThemeProvider theme={themes['light']}>
        <ChartStoreContext.Provider value={store as any}>{ui}</ChartStoreContext.Provider>
      </ThemeProvider>
    </StyledEngineProvider>
  );
}

function renderWithStore(ui: ReactNode, store: object) {
  return render(wrapWithStore(ui, store));
}

describe('group timeline charts', () => {
  it('refetches version timeline data when the group changes with the same duration', async () => {
    const alpha = makeGroup('group-alpha', 'Alpha');
    const beta = makeGroup('group-beta', 'Beta');
    const store = {
      getGroupVersionCountTimeline: vi.fn((_appID: string, groupID: string) =>
        Promise.resolve(
          groupID === alpha.id
            ? { '2026-01-01T00:00:00Z': { '1.0.0': 3 } }
            : { '2026-01-01T00:00:00Z': { '2.0.0': 5 } }
        )
      ),
    };

    const { rerender } = renderWithStore(
      <VersionCountTimeline group={alpha} duration={duration} isAnimationActive={false} />,
      store
    );

    await waitFor(() =>
      expect(store.getGroupVersionCountTimeline).toHaveBeenCalledWith('app-1', alpha.id, '1d')
    );
    await waitFor(() => expect(screen.getByText('1.0.0')).toBeTruthy());

    rerender(
      wrapWithStore(
        <VersionCountTimeline group={beta} duration={duration} isAnimationActive={false} />,
        store
      )
    );

    await waitFor(() =>
      expect(store.getGroupVersionCountTimeline).toHaveBeenCalledWith('app-1', beta.id, '1d')
    );
    await waitFor(() => expect(screen.getByText('2.0.0')).toBeTruthy());
    expect(screen.queryByText('1.0.0')).toBeNull();
  });

  it('ignores stale version timeline responses after the selected group changes', async () => {
    const alpha = makeGroup('group-alpha', 'Alpha');
    const beta = makeGroup('group-beta', 'Beta');
    const alphaTimeline = deferredTimeline<{ [key: string]: { [key: string]: number } }>();
    const store = {
      getGroupVersionCountTimeline: vi.fn((_appID: string, groupID: string) => {
        if (groupID === alpha.id) {
          return alphaTimeline.promise;
        }

        return Promise.resolve({ '2026-01-01T00:00:00Z': { '2.0.0': 5 } });
      }),
    };

    const { rerender } = renderWithStore(
      <VersionCountTimeline group={alpha} duration={duration} isAnimationActive={false} />,
      store
    );

    await waitFor(() =>
      expect(store.getGroupVersionCountTimeline).toHaveBeenCalledWith('app-1', alpha.id, '1d')
    );

    rerender(
      wrapWithStore(
        <VersionCountTimeline group={beta} duration={duration} isAnimationActive={false} />,
        store
      )
    );

    await waitFor(() =>
      expect(store.getGroupVersionCountTimeline).toHaveBeenCalledWith('app-1', beta.id, '1d')
    );
    await waitFor(() => expect(screen.getByText('2.0.0')).toBeTruthy());

    await act(async () => {
      alphaTimeline.resolve({ '2026-01-01T00:00:00Z': { '1.0.0': 3 } });
      await alphaTimeline.promise;
    });

    expect(screen.getByText('2.0.0')).toBeTruthy();
    expect(screen.queryByText('1.0.0')).toBeNull();
  });

  it('refetches status timeline data when the group changes with the same duration', async () => {
    const alpha = makeGroup('group-alpha', 'Alpha');
    const beta = makeGroup('group-beta', 'Beta');
    const store = {
      getGroupStatusCountTimeline: vi.fn((_appID: string, groupID: string) =>
        Promise.resolve(
          groupID === alpha.id
            ? { '2026-01-01T00:00:00Z': { 4: { '1.0.0': 3 } } }
            : { '2026-01-01T00:00:00Z': { 8: { '2.0.0': 5 } } }
        )
      ),
    };

    const { rerender } = renderWithStore(
      <StatusCountTimeline group={alpha} duration={duration} isAnimationActive={false} />,
      store
    );

    await waitFor(() =>
      expect(store.getGroupStatusCountTimeline).toHaveBeenCalledWith('app-1', alpha.id, '1d')
    );
    await waitFor(() => expect(screen.getByText('1.0.0')).toBeTruthy());

    rerender(
      wrapWithStore(
        <StatusCountTimeline group={beta} duration={duration} isAnimationActive={false} />,
        store
      )
    );

    await waitFor(() =>
      expect(store.getGroupStatusCountTimeline).toHaveBeenCalledWith('app-1', beta.id, '1d')
    );
    await waitFor(() => expect(screen.getByText('2.0.0')).toBeTruthy());
    expect(screen.queryByText('1.0.0')).toBeNull();
  });

  it('ignores stale status timeline responses after the selected group changes', async () => {
    const alpha = makeGroup('group-alpha', 'Alpha');
    const beta = makeGroup('group-beta', 'Beta');
    const alphaTimeline = deferredTimeline<{ [key: string]: { [key: number]: object } }>();
    const store = {
      getGroupStatusCountTimeline: vi.fn((_appID: string, groupID: string) => {
        if (groupID === alpha.id) {
          return alphaTimeline.promise;
        }

        return Promise.resolve({ '2026-01-01T00:00:00Z': { 8: { '2.0.0': 5 } } });
      }),
    };

    const { rerender } = renderWithStore(
      <StatusCountTimeline group={alpha} duration={duration} isAnimationActive={false} />,
      store
    );

    await waitFor(() =>
      expect(store.getGroupStatusCountTimeline).toHaveBeenCalledWith('app-1', alpha.id, '1d')
    );

    rerender(
      wrapWithStore(
        <StatusCountTimeline group={beta} duration={duration} isAnimationActive={false} />,
        store
      )
    );

    await waitFor(() =>
      expect(store.getGroupStatusCountTimeline).toHaveBeenCalledWith('app-1', beta.id, '1d')
    );
    await waitFor(() => expect(screen.getByText('2.0.0')).toBeTruthy());

    await act(async () => {
      alphaTimeline.resolve({ '2026-01-01T00:00:00Z': { 4: { '1.0.0': 3 } } });
      await alphaTimeline.promise;
    });

    expect(screen.getByText('2.0.0')).toBeTruthy();
    expect(screen.queryByText('1.0.0')).toBeNull();
  });
});
