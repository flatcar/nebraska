import { afterEach, beforeEach, describe, expect, it, MockInstance, vi } from 'vitest';

const mockStore = {
  getCachedApplication: vi.fn(),
  getApplication: vi.fn(),
  addChangeListener: vi.fn(),
  removeChangeListener: vi.fn(),
  setGroupUpdatesPaused: vi.fn(),
};

vi.mock('../../stores/Stores', () => ({
  applicationsStore: vi.fn(() => mockStore),
}));
// The charts load their own data and are not what these tests are about.
vi.mock('../../components/Instances/Charts', () => ({ default: () => null }));
vi.mock('../../components/Groups/GroupCharts/StatusCountTimeline', () => ({
  default: () => null,
}));
vi.mock('../../components/Groups/GroupCharts/VersionCountTimeline', () => ({
  default: () => null,
}));
vi.mock('../../components/common/TimeIntervalLinks', () => ({ default: () => null }));

import '../../i18n/config.ts';

import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { Provider } from 'react-redux';
import { MemoryRouter } from 'react-router';

import API from '../../api/API';
import { Group } from '../../api/apiDataTypes';
import GroupItemExtended from '../../components/Groups/GroupItemExtended';
import { makeGroup } from '../../TestHelpers/groups';
import { makeStore } from '../../TestHelpers/store';
import { theme } from '../../TestHelpers/theme';

async function renderGroupPage(group: Group, instanceMode?: string) {
  mockStore.getCachedApplication.mockReturnValue({ id: 'app-1', groups: [group] });
  const handlers = { handleUpdateGroup: vi.fn(), handleEditOverrides: vi.fn() };
  await act(async () => {
    render(
      <Provider store={makeStore(instanceMode)}>
        <StyledEngineProvider injectFirst>
          <ThemeProvider theme={theme}>
            <MemoryRouter initialEntries={['/apps/app-1/groups/group-1']}>
              <GroupItemExtended appID="app-1" groupID="group-1" {...handlers} />
            </MemoryRouter>
          </ThemeProvider>
        </StyledEngineProvider>
      </Provider>
    );
  });
  return handlers;
}

function panes() {
  return {
    defaults: screen.queryByTestId('default-policy-pane'),
    local: screen.queryByTestId('local-policy-pane'),
  };
}

describe('GroupItemExtended', () => {
  let confirm: MockInstance<typeof window.confirm>;

  beforeEach(() => {
    vi.spyOn(API, 'getGroupInstancesStats').mockResolvedValue({ total: 0 });
    mockStore.setGroupUpdatesPaused.mockReset().mockResolvedValue(undefined);
    confirm = vi.spyOn(window, 'confirm').mockReturnValue(true);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('keeps a single node page as it was', async () => {
    await renderGroupPage(makeGroup(null), 'single');

    expect(panes().defaults).toBeNull();
    expect(panes().local).toBeNull();
    expect(screen.getByText('Updates Timeout')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Open menu' })).toHaveProperty('disabled', false);
  });

  it('lets a control node edit the default and its own overrides', async () => {
    const handlers = await renderGroupPage(makeGroup(), 'control');
    const { defaults, local } = panes();

    expect(within(local!).getByRole('heading', { name: 'This Node (Control)' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Open menu' })).toHaveProperty('disabled', false);

    fireEvent.click(within(defaults!).getByRole('button', { name: 'Edit default policy' }));
    expect(handlers.handleUpdateGroup).toHaveBeenCalledWith('group-1', 'app-1');
    fireEvent.click(within(local!).getByRole('button', { name: "Edit this node's overrides" }));
    expect(handlers.handleEditOverrides).toHaveBeenCalledTimes(1);
  });

  it('offers an edge node its own overrides only', async () => {
    await renderGroupPage(makeGroup(), 'edge');
    const { defaults, local } = panes();

    expect(within(local!).getByRole('heading', { name: 'This Node (Edge)' })).toBeTruthy();
    expect(within(defaults!).queryByRole('button')).toBeNull();
    // An edge node refuses changes to the group itself.
    expect(screen.getByRole('button', { name: 'Open menu' })).toHaveProperty('disabled', true);
  });

  it('pauses updates on the node once confirmed', async () => {
    await renderGroupPage(makeGroup(), 'edge');

    await act(async () =>
      fireEvent.click(screen.getByRole('button', { name: 'Pause updates on this node' }))
    );

    expect(confirm).toHaveBeenCalledWith(expect.stringMatching(/^Pause updates on this node\?/));
    expect(mockStore.setGroupUpdatesPaused).toHaveBeenCalledWith('app-1', 'group-1', true);
  });

  it('resumes updates on the node once confirmed', async () => {
    await renderGroupPage(makeGroup({ policy_updates_enabled: false }), 'control');

    await act(async () =>
      fireEvent.click(screen.getByRole('button', { name: 'Resume updates on this node' }))
    );

    expect(confirm).toHaveBeenCalledWith(expect.stringMatching(/^Resume updates on this node\?/));
    expect(mockStore.setGroupUpdatesPaused).toHaveBeenCalledWith('app-1', 'group-1', false);
  });

  it('leaves the node as it is when the user cancels', async () => {
    confirm.mockReturnValue(false);
    await renderGroupPage(makeGroup(), 'control');

    await act(async () =>
      fireEvent.click(screen.getByRole('button', { name: 'Pause updates on this node' }))
    );

    expect(confirm).toHaveBeenCalledTimes(1);
    expect(mockStore.setGroupUpdatesPaused).not.toHaveBeenCalled();
  });

  it('tells the user when the node refuses', async () => {
    const alert = vi.spyOn(window, 'alert').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
    mockStore.setGroupUpdatesPaused.mockRejectedValue(new Response(null, { status: 500 }));
    await renderGroupPage(makeGroup(), 'control');

    await act(async () =>
      fireEvent.click(screen.getByRole('button', { name: 'Pause updates on this node' }))
    );

    expect(alert).toHaveBeenCalledWith(expect.stringMatching(/Something went wrong/));
  });
});
