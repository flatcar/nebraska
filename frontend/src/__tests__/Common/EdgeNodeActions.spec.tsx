import { beforeEach, describe, expect, it, vi } from 'vitest';

const mockStore = {
  getCachedApplications: vi.fn(),
  getCachedApplication: vi.fn(),
  getApplication: vi.fn(),
  getApplications: vi.fn(),
  addChangeListener: vi.fn(),
  removeChangeListener: vi.fn(),
  getPackageQueryParams: vi.fn(() => ({ page: 0, perPage: 10 })),
  getAndUpdatePackages: vi.fn(),
};

vi.mock('../../stores/Stores', () => ({
  applicationsStore: vi.fn(() => mockStore),
}));

import '../../i18n/config.ts';

import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { render, screen } from '@testing-library/react';
import React from 'react';
import { Provider } from 'react-redux';
import { MemoryRouter } from 'react-router';

import API from '../../api/API';
import { Application, Channel, Package } from '../../api/apiDataTypes';
import ApplicationList from '../../components/Applications/ApplicationList';
import ChannelList from '../../components/Channels/ChannelList';
import Header from '../../components/Header';
import PackageList from '../../components/Packages/List';
import { makeGroup } from '../../TestHelpers/groups';
import { makeStore } from '../../TestHelpers/store';
import { theme } from '../../TestHelpers/theme';

const pkg = {
  id: 'pkg-1',
  type: 1,
  version: '4081.2.0',
  url: 'https://update.release.flatcar-linux.net/',
  filename: 'flatcar_production_update.gz',
  description: null,
  size: null,
  hash: null,
  created_ts: '2026-01-01T00:00:00Z',
  channels_blacklist: null,
  application_id: 'app-1',
  arch: 1,
  extra_files: [],
} as Package;

const channel = {
  id: 'channel-1',
  name: 'stable',
  color: '#777777',
  created_ts: '2026-01-01T00:00:00Z',
  application_id: 'app-1',
  package_id: 'pkg-1',
  package: pkg,
  arch: 1,
} as unknown as Channel;

// An application as a node serves it. A control or edge node sends its groups
// with their local overrides, and a single node without.
function makeApp(fromDistributedNode: boolean) {
  return {
    id: 'app-1',
    name: 'Flatcar',
    description: 'Flatcar Container Linux',
    product_id: 'io.flatcar.linux',
    created_ts: '2026-01-01T00:00:00Z',
    groups: [makeGroup(fromDistributedNode ? {} : null)],
    channels: [channel],
    packages: { items: [pkg], totalCount: 1 },
    instances: { count: 0 },
  } as unknown as Application;
}

function renderWith(instanceMode: string | undefined, element: React.ReactElement) {
  render(
    <Provider store={makeStore(instanceMode)}>
      <StyledEngineProvider injectFirst>
        <ThemeProvider theme={theme}>
          <MemoryRouter>{element}</MemoryRouter>
        </ThemeProvider>
      </StyledEngineProvider>
    </Provider>
  );
}

function useApp(app: Application) {
  mockStore.getCachedApplications.mockReturnValue([app]);
  mockStore.getCachedApplication.mockReturnValue(app);
}

// The add button of the list, and the menus of its entries.
function adminActions() {
  return [
    screen.getByRole('button', { name: 'add' }),
    ...screen.getAllByRole('button', { name: 'Open menu' }),
  ];
}

const lists: [string, () => React.ReactElement][] = [
  ['applications', () => <ApplicationList />],
  ['channels', () => <ChannelList appID="app-1" />],
  ['packages', () => <PackageList appID="app-1" />],
];

describe('admin actions', () => {
  beforeEach(() => {
    Object.values(mockStore).forEach(mock => mock.mockClear());
    // The application cards count the instances of each group.
    vi.spyOn(API, 'getInstancesCount').mockResolvedValue(0);
  });

  it.each(lists)('are unavailable for %s on an edge node', (_, list) => {
    useApp(makeApp(true));
    renderWith('edge', list());

    adminActions().forEach(button => expect(button).toHaveProperty('disabled', true));
  });

  it.each(lists)('stay available for %s on a control node', (_, list) => {
    useApp(makeApp(true));
    renderWith('control', list());

    adminActions().forEach(button => expect(button).toHaveProperty('disabled', false));
  });

  it.each(lists)('stay available for %s on a single node', (_, list) => {
    useApp(makeApp(false));
    renderWith('single', list());

    adminActions().forEach(button => expect(button).toHaveProperty('disabled', false));
  });
});

describe('the role of the node', () => {
  it.each([
    ['control', 'Control node'],
    ['edge', 'Edge node'],
  ])('is named in the header of a %s node', (instanceMode, label) => {
    renderWith(instanceMode, <Header />);
    expect(screen.getByTestId('node-mode').textContent).toBe(label);
  });

  it.each(['single', undefined])('is not named in the header of a %s node', instanceMode => {
    renderWith(instanceMode, <Header />);
    expect(screen.queryByTestId('node-mode')).toBeNull();
  });
});
