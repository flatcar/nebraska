import { ThemeProvider } from '@mui/material/styles';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { describe, expect, it } from 'vitest';

import { Group } from '../../api/apiDataTypes';
import VersionCountTimeline from '../../components/Groups/GroupCharts/VersionCountTimeline';
import themes from '../../lib/themes';
import GroupChartsStore from '../../stores/GroupChartsStore';
import { groupChartStoreContext } from '../../stores/Stores';

describe('VersionCountTimeline Component', () => {
  const mockGroup: Group = {
    id: 'test-group-id',
    name: 'Test Group',
    description: 'Test Group Description',
    created_ts: '2026-01-01T00:00:00Z',
    rollout_in_progress: false,
    application_id: 'test-app-id',
    channel_id: 'test-channel-id',
    policy_updates_enabled: true,
    policy_safe_mode: false,
    policy_office_hours: false,
    policy_timezone: 'UTC',
    policy_period_interval: '15m',
    policy_max_updates_per_period: 10,
    policy_update_timeout: '1h',
    channel: {
      id: 'test-channel-id',
      name: 'test-channel',
      color: '#14b9d6',
      created_ts: '2026-01-01T00:00:00Z',
      application_id: 'test-app-id',
      package_id: 'test-pkg-id',
      package: {
        id: 'test-pkg-id',
        type: 1,
        version: '3510.2.0+test',
        url: '',
        filename: '',
        description: '',
        size: '100',
        hash: '',
        created_ts: '2026-01-01T00:00:00Z',
        channels_blacklist: null,
        application_id: 'test-app-id',
        arch: 1,
        extra_files: [],
      },
      arch: 1,
    },
    track: 'test-channel',
  };

  const duration = { displayValue: '1 day', queryValue: '1d', disabled: false };

  it('renders build metadata versions with correct count and percentage in table', async () => {
    const mockTimeline = {
      '2026-08-15T00:00:00Z': {
        '3510.2.0+test': 7,
        '2191.5.0': 3,
      },
      '2026-08-15T01:00:00Z': {
        '3510.2.0+test': 7,
        '2191.5.0': 3,
      },
    };

    class GroupChartsStoreMock extends GroupChartsStore {
      async getGroupVersionCountTimeline() {
        return mockTimeline;
      }
    }

    const ChartStoreContext = groupChartStoreContext();

    render(
      <ChartStoreContext.Provider value={new GroupChartsStoreMock()}>
        <MemoryRouter>
          <ThemeProvider theme={themes['light']}>
            <VersionCountTimeline group={mockGroup} duration={duration} isAnimationActive={false} />
          </ThemeProvider>
        </MemoryRouter>
      </ChartStoreContext.Provider>
    );

    // Wait for the version table to render with the raw version containing build metadata
    await waitFor(() => {
      expect(screen.getByText('3510.2.0+test')).toBeTruthy();
    });

    expect(screen.getByText('2191.5.0')).toBeTruthy();
    expect(screen.getAllByText('7').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('3').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('70.0')).toBeTruthy();
    expect(screen.getByText('30.0')).toBeTruthy();
  });

  it('distinguishes multiple metadata variants of the same core version', async () => {
    const mockTimeline = {
      '2026-08-15T00:00:00Z': {
        '1.2.3+aws': 4,
        '1.2.3+azure': 6,
      },
    };

    class GroupChartsStoreMock extends GroupChartsStore {
      async getGroupVersionCountTimeline() {
        return mockTimeline;
      }
    }

    const ChartStoreContext = groupChartStoreContext();

    render(
      <ChartStoreContext.Provider value={new GroupChartsStoreMock()}>
        <MemoryRouter>
          <ThemeProvider theme={themes['light']}>
            <VersionCountTimeline group={mockGroup} duration={duration} isAnimationActive={false} />
          </ThemeProvider>
        </MemoryRouter>
      </ChartStoreContext.Provider>
    );

    await waitFor(() => {
      expect(screen.getByText('1.2.3+aws')).toBeTruthy();
      expect(screen.getByText('1.2.3+azure')).toBeTruthy();
    });

    expect(screen.getAllByText('4').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('6').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('40.0')).toBeTruthy();
    expect(screen.getByText('60.0')).toBeTruthy();
  });
});
