import '../../i18n/config.ts';

import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { act, fireEvent, render, screen } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router';
import { describe, expect, it, vi } from 'vitest';

import ModalButton from '../../components/common/ModalButton';
import MoreMenu from '../../components/common/MoreMenu';
import { theme } from '../../TestHelpers/theme';

function wrap(element: React.ReactElement) {
  return render(
    <StyledEngineProvider injectFirst>
      <ThemeProvider theme={theme}>
        <MemoryRouter>{element}</MemoryRouter>
      </ThemeProvider>
    </StyledEngineProvider>
  );
}

const reason = 'Only available on a control or single node';

// An action a node refuses stays in sight, disabled, and says on hover where
// it can be made.
describe('unavailable actions', () => {
  it.each([
    [
      'menu',
      () => <MoreMenu options={[{ label: 'Edit', action: vi.fn() }]} disabledReason={reason} />,
    ],
    [
      'add button',
      () => <ModalButton modalToOpen="AddPackageModal" data={{}} disabledReason={reason} />,
    ],
  ])('disables the %s and explains why on hover', async (_, element) => {
    wrap(element());
    const button = screen.getByRole('button');
    expect(button).toHaveProperty('disabled', true);

    // The disabled button fires no events, so its wrapper shows the reason.
    await act(async () => fireEvent.mouseOver(button.parentElement!));
    expect((await screen.findByRole('tooltip')).textContent).toBe(reason);
  });

  it.each([
    ['menu', () => <MoreMenu options={[{ label: 'Edit', action: vi.fn() }]} />],
    ['add button', () => <ModalButton modalToOpen="AddPackageModal" data={{}} />],
  ])('leaves the %s as it was otherwise', async (_, element) => {
    wrap(element());
    const button = screen.getByRole('button');
    expect(button).toHaveProperty('disabled', false);

    await act(async () => fireEvent.mouseOver(button));
    expect(screen.queryByRole('tooltip')).toBeNull();
  });

  it('keeps a disabled menu closed', async () => {
    const action = vi.fn();
    wrap(<MoreMenu options={[{ label: 'Edit', action }]} disabledReason={reason} />);

    await act(async () => fireEvent.click(screen.getByRole('button')));
    expect(screen.queryByRole('menuitem', { name: 'Edit' })).toBeNull();
    expect(action).not.toHaveBeenCalled();
  });

  it('keeps a menu its caller disables', () => {
    wrap(
      <MoreMenu
        options={[{ label: 'Edit', action: vi.fn() }]}
        iconButtonProps={{ disabled: true }}
      />
    );
    expect(screen.getByRole('button')).toHaveProperty('disabled', true);
  });
});
