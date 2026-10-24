import { describe, expect, it, vi } from 'vitest';

vi.mock('../../stores/Stores', () => ({
  applicationsStore: vi.fn(),
}));

import '../../i18n/config.ts';

import { fireEvent, render, screen, within } from '@testing-library/react';

import { Channel } from '../../api/apiDataTypes';
import EditDialog from '../../components/Packages/EditDialog';

const channel: Channel = {
  id: 'channel-1',
  name: 'stable',
  color: '',
  created_ts: '',
  application_id: 'app-1',
  package_id: null,
  package: null,
  arch: 1,
};

describe('Packages EditDialog', () => {
  it('should keep the dialog when the channels blacklist is closed with Escape', () => {
    render(
      <EditDialog
        create
        show
        onHide={() => {}}
        data={{ appID: 'app-1', channels: [channel], package: undefined as any }}
      />
    );
    const select = screen
      .getAllByRole('combobox')
      .find(el => el.closest('.MuiFormControl-root')?.textContent?.includes('Channels Blacklist'))!;
    fireEvent.mouseDown(select);
    const option = within(screen.getByRole('listbox')).getByRole('option', { name: 'stable' });
    fireEvent.keyDown(option, { key: 'Escape' });
    expect(screen.queryByRole('listbox')).toBeNull();
    expect(screen.getByText('Channels Blacklist')).toBeTruthy();
  });
});
