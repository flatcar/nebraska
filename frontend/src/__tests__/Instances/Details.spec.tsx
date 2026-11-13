import { beforeEach, describe, expect, it, vi } from 'vitest';

const { mockGetInstanceStatusHistory, mockUpdateInstance } = vi.hoisted(() => ({
  mockGetInstanceStatusHistory: vi.fn(),
  mockUpdateInstance: vi.fn(),
}));

vi.mock('../../api/API', () => ({
  default: {
    getInstanceStatusHistory: mockGetInstanceStatusHistory,
    updateInstance: mockUpdateInstance,
  },
}));

import '../../i18n/config.ts';

import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { MemoryRouter } from 'react-router';

import type { Application, Group, Instance } from '../../api/apiDataTypes';
import DetailsView from '../../components/Instances/Details';
import { theme } from '../../TestHelpers/theme';

const instance = {
  id: 'instance-1',
  alias: '',
  created_ts: '2026-01-01T00:00:00Z',
  ip: '10.0.0.1',
  application: {
    instance_id: 'instance-1',
    application_id: 'app-1',
    group_id: 'group-1',
    version: '4152.2.3',
    created_ts: '2026-01-01T00:00:00Z',
    status: 2,
    last_check_for_updates: '2026-01-01T00:00:00Z',
  },
} as unknown as Instance;

const application = { id: 'app-1', name: 'Flatcar Container Linux' } as unknown as Application;
const group = { id: 'group-1', name: 'Alpha', application_id: 'app-1' } as unknown as Group;

function renderDetails() {
  const onInstanceUpdated = vi.fn();
  render(
    <StyledEngineProvider injectFirst>
      <ThemeProvider theme={theme}>
        <MemoryRouter>
          <DetailsView
            application={application}
            group={group}
            instance={instance}
            onInstanceUpdated={onInstanceUpdated}
          />
        </MemoryRouter>
      </ThemeProvider>
    </StyledEngineProvider>
  );
  return onInstanceUpdated;
}

async function openRenameDialog() {
  await waitFor(() => expect(mockGetInstanceStatusHistory).toHaveBeenCalledTimes(1));
  fireEvent.click(screen.getByTestId('more-menu-open-button'));
  fireEvent.click(screen.getByTestId('more-menu-item'));
  return screen.findByTestId('instance-edit-form');
}

describe('DetailsView rename dialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetInstanceStatusHistory.mockResolvedValue([]);
    mockUpdateInstance.mockResolvedValue({ ...instance, alias: 'renamed' });
  });

  it('does not report an update when the dialog is cancelled', async () => {
    const onInstanceUpdated = renderDetails();
    await openRenameDialog();

    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));

    expect(mockUpdateInstance).not.toHaveBeenCalled();
    expect(onInstanceUpdated).not.toHaveBeenCalled();
    expect(mockGetInstanceStatusHistory).toHaveBeenCalledTimes(1);
  });

  it('reports an update after a successful rename', async () => {
    const onInstanceUpdated = renderDetails();
    const form = await openRenameDialog();

    fireEvent.change(within(form).getByRole('textbox'), { target: { value: 'renamed' } });
    fireEvent.click(screen.getByRole('button', { name: /save/i }));

    await waitFor(() => expect(mockUpdateInstance).toHaveBeenCalledWith('instance-1', 'renamed'));
    await waitFor(() => expect(onInstanceUpdated).toHaveBeenCalledTimes(1));
  });
});
