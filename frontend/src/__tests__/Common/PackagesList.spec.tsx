import { beforeEach, describe, expect, it, vi } from 'vitest';

const storeMocks = vi.hoisted(() => ({
  getCachedApplication: vi.fn(),
  getApplication: vi.fn(),
  addChangeListener: vi.fn(),
  removeChangeListener: vi.fn(),
  getPackageQueryParams: vi.fn(() => ({ page: 0, perPage: 10 })),
  getAndUpdatePackages: vi.fn(),
  updatePackage: vi.fn(),
}));

vi.mock('../../stores/Stores', () => ({
  applicationsStore: vi.fn(() => storeMocks),
}));

vi.mock('../../api/API', async () => ({
  ...(await vi.importActual('../../api/API')),
  getPackages: vi.fn().mockResolvedValue({ packages: [], totalCount: 0 }),
  getApplication: vi.fn().mockResolvedValue({}),
}));

import '../../i18n/config.ts';

import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { fireEvent, render, screen } from '@testing-library/react';

import List from '../../components/Packages/List.tsx';
import themes from '../../lib/themes';

vi.mock('../../components/Packages/Item.tsx', () => ({
  default: ({ packageItem, handleUpdatePackage }: any) => (
    <button onClick={() => handleUpdatePackage(packageItem.id)}>
      {packageItem.extra_files?.[0]?.name || 'no extra files'}
    </button>
  ),
}));

vi.mock('../../components/Packages/EditDialog.tsx', () => ({
  default: ({ data, onPackageUpdated }: any) =>
    data.package?.id ? (
      <button
        onClick={() =>
          onPackageUpdated({
            id: 'pkg1',
            application_id: 'app123',
            extra_files: [{ name: 'signature.txt' }],
          })
        }
      >
        Save package
      </button>
    ) : null,
}));

describe('List Component', () => {
  const minProps = {
    appID: 'app123',
  };

  beforeEach(() => {
    storeMocks.getCachedApplication.mockReset();
    storeMocks.getApplication.mockReset();
    storeMocks.addChangeListener.mockReset();
    storeMocks.removeChangeListener.mockReset();
    storeMocks.getPackageQueryParams.mockReturnValue({ page: 0, perPage: 10 });
    storeMocks.getAndUpdatePackages.mockReset();
    storeMocks.updatePackage.mockReset();
  });

  it('should render the list', () => {
    storeMocks.getCachedApplication.mockReturnValueOnce(null);

    render(
      <StyledEngineProvider injectFirst>
        <ThemeProvider theme={themes['light']}>
          <List {...minProps} />
        </ThemeProvider>
      </StyledEngineProvider>
    );
    expect(screen.getByText('Packages')).toBeTruthy();
  });

  it('updates edited package data locally without sending a second update request', () => {
    storeMocks.getCachedApplication.mockReturnValue({
      id: 'app123',
      channels: [],
      packages: {
        totalCount: 1,
        items: [{ id: 'pkg1', application_id: 'app123', extra_files: [] }],
      },
    });

    render(
      <StyledEngineProvider injectFirst>
        <ThemeProvider theme={themes['light']}>
          <List {...minProps} />
        </ThemeProvider>
      </StyledEngineProvider>
    );

    fireEvent.click(screen.getByText('no extra files'));
    fireEvent.click(screen.getByText('Save package'));

    expect(screen.getByText('signature.txt')).toBeTruthy();
    expect(storeMocks.updatePackage).not.toHaveBeenCalled();
  });
});
