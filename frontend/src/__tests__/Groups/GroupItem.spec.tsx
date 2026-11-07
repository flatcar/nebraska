import '../../i18n/config.ts';

import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { render, screen, within } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { describe, expect, it, vi } from 'vitest';

import { Group } from '../../api/apiDataTypes';
import { PureGroupItem } from '../../components/Groups/GroupItem';
import { makeGroup } from '../../TestHelpers/groups';
import { theme } from '../../TestHelpers/theme';
import { InstanceMode } from '../../utils/instanceMode';

function renderCard(group: Group, instanceMode?: InstanceMode) {
  render(
    <StyledEngineProvider injectFirst>
      <ThemeProvider theme={theme}>
        <MemoryRouter>
          <PureGroupItem
            group={group}
            versionBreakdown={[]}
            totalInstances={0}
            handleUpdateGroup={vi.fn()}
            deleteGroup={vi.fn()}
            instanceMode={instanceMode}
          />
        </MemoryRouter>
      </ThemeProvider>
    </StyledEngineProvider>
  );
  return screen.getByTestId('list-item');
}

describe('PureGroupItem', () => {
  it('shows what applies on this node and marks overrides', () => {
    const card = renderCard(
      makeGroup({ policy_updates_enabled: false, policy_max_updates_per_period: 5 }),
      'edge'
    );

    expect(card.textContent).toContain('Paused on this node');
    expect(card.textContent).toContain('Max 5 / 15 minutes');
    expect(within(card).getAllByText('Override')).toHaveLength(2);
  });

  it('shows updates the default disables as disabled', () => {
    const card = renderCard({ ...makeGroup(), policy_updates_enabled: false }, 'edge');

    expect(card.textContent).toContain('Disabled');
    expect(card.textContent).not.toContain('Paused');
    expect(within(card).queryByText('Override')).toBeNull();
  });

  it('leaves a group without overrides as it was', () => {
    const card = renderCard(makeGroup(null));

    expect(card.textContent).toContain('Enabled');
    expect(card.textContent).toContain('Max 2 / 15 minutes');
    expect(within(card).queryByText('Override')).toBeNull();
    expect(within(card).getByTestId('more-menu-open-button')).toHaveProperty('disabled', false);
  });

  it('shows group changes as unavailable on an edge node', () => {
    const card = renderCard(makeGroup(), 'edge');
    expect(within(card).getByTestId('more-menu-open-button')).toHaveProperty('disabled', true);
  });

  it('offers group changes on a control node', () => {
    const card = renderCard(makeGroup(), 'control');
    expect(within(card).getByTestId('more-menu-open-button')).toHaveProperty('disabled', false);
  });
});
