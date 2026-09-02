import '../../i18n/config.ts';

import { ThemeProvider } from '@mui/material/styles';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import API from '../../api/API';
import { Application, Group, Instance } from '../../api/apiDataTypes';
import DetailsView from '../../components/Instances/Details';
import themes from '../../lib/themes';

function renderDetails(onInstanceUpdated = vi.fn()) {
  const application = {
    id: 'app-1',
    name: 'Test App',
  } as Application;

  const group = {
    id: 'group-1',
    name: 'Test Group',
    channel: null,
  } as unknown as Group;

  const instance = {
    id: 'instance-1',
    alias: '',
    ip: '10.0.0.1',
    application: {
      instance_id: 'instance-1',
      application_id: 'app-1',
      group_id: 'group-1',
      version: '1.0.0',
      status: 3,
      last_check_for_updates: '2024-01-01T00:00:00Z',
    },
    statusInfo: {
      type: 'complete',
      bgColor: '#000',
      textColor: '#fff',
      explanation: 'Up to date',
    },
  } as Instance;

  return {
    onInstanceUpdated,
    ...render(
      <BrowserRouter>
        <ThemeProvider theme={themes['light']}>
          <DetailsView
            application={application}
            group={group}
            instance={instance}
            onInstanceUpdated={onInstanceUpdated}
          />
        </ThemeProvider>
      </BrowserRouter>
    ),
  };
}

async function openEditDialog() {
  await waitFor(() => {
    expect(screen.queryByRole('progressbar')).toBeNull();
  });

  await act(async () => {
    fireEvent.click(screen.getByTestId('more-menu-open-button'));
  });

  await act(async () => {
    fireEvent.click(screen.getByText('Rename'));
  });

  await waitFor(() => {
    expect(screen.getByTestId('instance-edit-form')).toBeTruthy();
  });
}

describe('DetailsView edit dialog', () => {
  beforeEach(() => {
    vi.spyOn(API, 'getInstanceStatusHistory').mockResolvedValue([]);
  });

  it('does not refetch instance data when edit is canceled', async () => {
    const onInstanceUpdated = vi.fn();
    renderDetails(onInstanceUpdated);

    await openEditDialog();

    await act(async () => {
      fireEvent.click(screen.getByText('Cancel'));
    });

    expect(onInstanceUpdated).not.toHaveBeenCalled();
  });

  it('refetches instance data after a successful save', async () => {
    const updatedInstance = {
      id: 'instance-1',
      alias: 'renamed-instance',
    } as Instance;

    vi.spyOn(API, 'updateInstance').mockResolvedValue(updatedInstance);

    const onInstanceUpdated = vi.fn();
    renderDetails(onInstanceUpdated);

    await openEditDialog();

    await act(async () => {
      fireEvent.change(screen.getByLabelText('Name'), {
        target: { value: 'renamed-instance' },
      });
      fireEvent.click(screen.getByText('Save'));
    });

    await waitFor(() => {
      expect(onInstanceUpdated).toHaveBeenCalledTimes(1);
    });
  });
});
