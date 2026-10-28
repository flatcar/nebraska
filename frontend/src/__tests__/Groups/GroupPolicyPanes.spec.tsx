import '../../i18n/config.ts';

import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { cleanup, fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { Group } from '../../api/apiDataTypes';
import GroupPolicyPanes from '../../components/Groups/GroupPolicyPanes';
import { makeGroup } from '../../TestHelpers/groups';
import { theme } from '../../TestHelpers/theme';

function renderPanes(group: Group, instanceMode: 'control' | 'edge') {
  const handlers = {
    onEditDefault: vi.fn(),
    onEditOverrides: vi.fn(),
    onPauseUpdates: vi.fn(),
    onResumeUpdates: vi.fn(),
  };
  render(
    <StyledEngineProvider injectFirst>
      <ThemeProvider theme={theme}>
        <GroupPolicyPanes group={group} instanceMode={instanceMode} {...handlers} />
      </ThemeProvider>
    </StyledEngineProvider>
  );
  return {
    handlers,
    defaults: screen.getByTestId('default-policy-pane'),
    local: screen.getByTestId('local-policy-pane'),
  };
}

function rowValue(pane: HTMLElement, label: string) {
  const row = within(pane)
    .getAllByTestId('policy-row')
    .find(r => r.firstElementChild?.textContent === label);
  if (!row) {
    throw new Error(`no ${label} row`);
  }
  return row;
}

describe('GroupPolicyPanes', () => {
  it('lets a control node edit both the default and its own overrides', () => {
    const { handlers, defaults, local } = renderPanes(makeGroup(), 'control');

    expect(within(defaults).getByRole('heading', { name: 'Default Policy' })).toBeTruthy();
    expect(within(local).getByRole('heading', { name: 'This Node (Control)' })).toBeTruthy();

    fireEvent.click(within(defaults).getByRole('button', { name: 'Edit default policy' }));
    expect(handlers.onEditDefault).toHaveBeenCalledTimes(1);

    fireEvent.click(within(local).getByRole('button', { name: "Edit this node's overrides" }));
    expect(handlers.onEditOverrides).toHaveBeenCalledTimes(1);
  });

  it('shows an edge node the default without a way to edit it', () => {
    const { defaults, local } = renderPanes(makeGroup(), 'edge');

    expect(within(defaults).queryByRole('button')).toBeNull();
    expect(within(defaults).getByTitle(/Read-only on an edge node/)).toBeTruthy();
    expect(within(local).getByRole('heading', { name: 'This Node (Edge)' })).toBeTruthy();
    expect(within(local).getByRole('button', { name: "Edit this node's overrides" })).toBeTruthy();
  });

  it('shows what applies on the node and marks its overrides', () => {
    const group = makeGroup({
      policy_updates_enabled: false,
      policy_max_updates_per_period: 5,
      policy_period_interval: '2 hours',
    });
    const { defaults, local } = renderPanes(group, 'edge');

    expect(rowValue(defaults, 'Updates').textContent).toContain('Enabled');
    expect(rowValue(defaults, 'Updates Policy').textContent).toContain('Max 2 / 15 minutes');
    expect(within(defaults).queryByText('Override')).toBeNull();

    expect(rowValue(local, 'Updates').textContent).toContain('Paused on this node');
    expect(rowValue(local, 'Updates Policy').textContent).toContain('Max 5 / 2 hours');
    expect(within(rowValue(local, 'Updates')).getByText('Override')).toBeTruthy();
    expect(within(rowValue(local, 'Updates Policy')).getByText('Override')).toBeTruthy();
    expect(within(rowValue(local, 'Safe Mode')).queryByText('Override')).toBeNull();
  });

  it('tells a node paused on its own from updates the default disables', () => {
    const disabled = { ...makeGroup(), policy_updates_enabled: false };
    const byDefault = renderPanes(disabled, 'edge');

    expect(rowValue(byDefault.defaults, 'Updates').textContent).toContain('Disabled');
    expect(rowValue(byDefault.local, 'Updates').textContent).toContain('Disabled');
    expect(rowValue(byDefault.local, 'Updates').textContent).not.toContain('Paused');
    cleanup();

    const pausedToo = renderPanes(
      {
        ...disabled,
        local_overrides: makeGroup({ policy_updates_enabled: false }).local_overrides,
      },
      'edge'
    );
    expect(rowValue(pausedToo.defaults, 'Updates').textContent).toContain('Disabled');
    expect(rowValue(pausedToo.local, 'Updates').textContent).toContain('Paused on this node');
  });

  it.each(['control', 'edge'] as const)('lets a %s node pause itself', instanceMode => {
    const { handlers, defaults, local } = renderPanes(makeGroup(), instanceMode);

    fireEvent.click(within(local).getByRole('button', { name: 'Pause updates on this node' }));
    expect(handlers.onPauseUpdates).toHaveBeenCalledTimes(1);
    expect(within(local).queryByRole('button', { name: /Resume/ })).toBeNull();
    expect(within(defaults).queryByRole('button', { name: /Pause/ })).toBeNull();
  });

  it.each(['control', 'edge'] as const)('lets a paused %s node resume', instanceMode => {
    const { handlers, local } = renderPanes(
      makeGroup({ policy_updates_enabled: false }),
      instanceMode
    );

    fireEvent.click(within(local).getByRole('button', { name: 'Resume updates on this node' }));
    expect(handlers.onResumeUpdates).toHaveBeenCalledTimes(1);
    expect(within(local).queryByRole('button', { name: /Pause/ })).toBeNull();
  });

  it('shows the office hours timezone', () => {
    const group = makeGroup({ policy_office_hours: true, policy_timezone: 'Asia/Tokyo' });
    const { defaults, local } = renderPanes(group, 'control');

    expect(rowValue(defaults, 'Only Office Hours').textContent).toContain('No');
    expect(rowValue(local, 'Only Office Hours').textContent).toContain('Yes (Asia/Tokyo)');
  });
});
