import { beforeEach, describe, expect, it, vi } from 'vitest';

const mockStore = {
  getCachedApplication: vi.fn(),
  addChangeListener: vi.fn(),
  removeChangeListener: vi.fn(),
};

vi.mock('../../stores/Stores', () => ({
  applicationsStore: vi.fn(() => mockStore),
}));
// Each card loads its own data and is not what these tests are about.
vi.mock('../../components/Groups/GroupItem', () => ({ default: () => <li>group</li> }));

import '../../i18n/config.ts';

import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { render, screen } from '@testing-library/react';
import { Provider } from 'react-redux';
import { MemoryRouter } from 'react-router';

import GroupList from '../../components/Groups/GroupList';
import { makeGroup } from '../../TestHelpers/groups';
import { makeStore } from '../../TestHelpers/store';
import { theme } from '../../TestHelpers/theme';

// A single node sends its groups without local overrides.
function renderList(instanceMode: string | undefined, groups = [makeGroup()]) {
  mockStore.getCachedApplication.mockReturnValue({ id: 'app-1', groups, channels: [] });
  render(
    <Provider store={makeStore(instanceMode)}>
      <StyledEngineProvider injectFirst>
        <ThemeProvider theme={theme}>
          <MemoryRouter>
            <GroupList appID="app-1" />
          </MemoryRouter>
        </ThemeProvider>
      </StyledEngineProvider>
    </Provider>
  );
}

describe('GroupList', () => {
  beforeEach(() => {
    mockStore.getCachedApplication.mockReset();
  });

  it('lets a single node add a group', () => {
    renderList('single', [makeGroup(null)]);
    expect(screen.getByRole('button', { name: 'add' })).toHaveProperty('disabled', false);
  });

  it('lets a control node add a group', () => {
    renderList('control');
    expect(screen.getByRole('button', { name: 'add' })).toHaveProperty('disabled', false);
  });

  it('shows adding a group as unavailable on an edge node, which refuses it', () => {
    renderList('edge');
    expect(screen.getByText('group')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'add' })).toHaveProperty('disabled', true);
  });

  it('keeps an older server that reports no mode a single node', () => {
    renderList(undefined, [makeGroup(null)]);
    expect(screen.getByRole('button', { name: 'add' })).toHaveProperty('disabled', false);
  });
});
