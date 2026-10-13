import { beforeEach, describe, expect, it, vi } from 'vitest';

const mockUpdateGroup = vi.fn();
const mockCreateGroup = vi.fn();

vi.mock('../../stores/Stores', () => ({
  applicationsStore: vi.fn(() => ({
    updateGroup: mockUpdateGroup,
    createGroup: mockCreateGroup,
  })),
}));

import '../../i18n/config.ts';

import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';

import GroupEditDialog from '../../components/Groups/GroupEditDialog';
import { theme } from '../../TestHelpers/theme';
import { MAX_UPDATES_PER_TIME_PERIOD } from '../../utils/helpers';

function makeGroup(maxUpdates: number, periodInterval: string, safeMode = false) {
  return {
    id: 'group-1',
    name: 'Stable',
    track: 'stable',
    description: '',
    policy_timezone: 'UTC',
    policy_updates_enabled: true,
    policy_office_hours: false,
    policy_safe_mode: safeMode,
    policy_max_updates_per_period: maxUpdates,
    policy_period_interval: periodInterval,
    policy_update_timeout: '60 minutes',
    channel: null,
  };
}

async function renderPolicyTab(group?: ReturnType<typeof makeGroup>) {
  render(
    <StyledEngineProvider injectFirst>
      <ThemeProvider theme={theme}>
        <MemoryRouter initialEntries={['/app/123']}>
          <GroupEditDialog
            create={!group}
            data={group ? { group, channels: [] } : { channels: [] }}
            onHide={() => {}}
            show
          />
        </MemoryRouter>
      </ThemeProvider>
    </StyledEngineProvider>
  );
  await act(async () => fireEvent.click(screen.getByRole('tab', { name: 'Policy' })));
}

function unlimitedSwitch() {
  return screen.getByRole('switch', { name: 'Unlimited updates' });
}

async function chooseUnlimited() {
  if (!(unlimitedSwitch() as HTMLInputElement).checked) {
    await act(async () => fireEvent.click(unlimitedSwitch()));
  }
}

async function chooseLimited() {
  if ((unlimitedSwitch() as HTMLInputElement).checked) {
    await act(async () => fireEvent.click(unlimitedSwitch()));
  }
}

function maxUpdatesInput() {
  return screen.queryByRole('spinbutton', { name: 'Max number of updates' }) as HTMLInputElement;
}

async function setMaxUpdates(value: string) {
  await act(async () => fireEvent.change(maxUpdatesInput(), { target: { value } }));
}

async function submit(name = 'Save') {
  await act(async () => fireEvent.click(screen.getByRole('button', { name })));
}

describe('GroupEditDialog update limits', () => {
  beforeEach(() => {
    mockUpdateGroup.mockReset().mockResolvedValue({});
    mockCreateGroup.mockReset().mockResolvedValue({});
  });

  it('saves a limited group as unlimited', async () => {
    await renderPolicyTab(makeGroup(5, '2 hours'));
    expect(maxUpdatesInput().value).toBe('5');

    await chooseUnlimited();
    await waitFor(() => expect(maxUpdatesInput()).toBeNull());
    await submit();

    await waitFor(() => expect(mockUpdateGroup).toHaveBeenCalledTimes(1));
    const saved = mockUpdateGroup.mock.calls[0][0];
    expect(saved.policy_max_updates_per_period).toBe(MAX_UPDATES_PER_TIME_PERIOD);
    expect(saved.policy_period_interval).toBe('2 hours');
  });

  it('keeps the stored period when saving a group that stays unlimited', async () => {
    await renderPolicyTab(makeGroup(MAX_UPDATES_PER_TIME_PERIOD, '15 minutes', true));
    await submit();

    await waitFor(() => expect(mockUpdateGroup).toHaveBeenCalledTimes(1));
    const saved = mockUpdateGroup.mock.calls[0][0];
    expect(saved.policy_max_updates_per_period).toBe(MAX_UPDATES_PER_TIME_PERIOD);
    expect(saved.policy_period_interval).toBe('15 minutes');
  });

  it('explains that safe mode still applies to unlimited groups', async () => {
    await renderPolicyTab(makeGroup(MAX_UPDATES_PER_TIME_PERIOD, '1 minutes', true));
    expect(screen.getByText(/safe mode still updates 1 instance at a time/)).toBeTruthy();
    expect(screen.queryByText(/can all update in the same period/)).toBeNull();
  });

  it('switches an unlimited group to a safe default limit', async () => {
    await renderPolicyTab(makeGroup(MAX_UPDATES_PER_TIME_PERIOD, '1 minutes'));
    expect(maxUpdatesInput()).toBeNull();

    await chooseLimited();
    expect(maxUpdatesInput().value).toBe('1');
    await submit();

    await waitFor(() => expect(mockUpdateGroup).toHaveBeenCalledTimes(1));
    const saved = mockUpdateGroup.mock.calls[0][0];
    expect(saved.policy_max_updates_per_period).toBe(1);
    expect(saved.policy_period_interval).toBe('1 hours');
  });

  it('restores the previous limit when switching back from unlimited', async () => {
    await renderPolicyTab(makeGroup(5, '2 hours'));
    await setMaxUpdates('7');

    await chooseUnlimited();
    await chooseLimited();

    expect(maxUpdatesInput().value).toBe('7');
  });

  it('rejects limits that the backend would treat as unlimited', async () => {
    await renderPolicyTab(makeGroup(5, '2 hours'));
    await setMaxUpdates(String(MAX_UPDATES_PER_TIME_PERIOD));
    await submit();

    expect(await screen.findByText(/Must be less than 900,000/)).toBeTruthy();
    expect(mockUpdateGroup).not.toHaveBeenCalled();
  });

  it('ignores an invalid hidden limit when creating an unlimited group', async () => {
    await renderPolicyTab();
    await act(async () => fireEvent.click(screen.getByRole('tab', { name: 'Details' })));
    await act(async () =>
      fireEvent.change(screen.getByRole('textbox', { name: /Name/ }), {
        target: { value: 'Canary' },
      })
    );
    await act(async () => fireEvent.click(screen.getByRole('tab', { name: 'Policy' })));
    await setMaxUpdates('');

    await chooseUnlimited();
    await submit('add');

    await waitFor(() => expect(mockCreateGroup).toHaveBeenCalledTimes(1));
    expect(mockCreateGroup.mock.calls[0][0].policy_max_updates_per_period).toBe(
      MAX_UPDATES_PER_TIME_PERIOD
    );
  });
});
