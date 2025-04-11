import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../../stores/Stores', () => ({
  applicationsStore: vi.fn(() => ({
    getCachedApplication: vi.fn(),
    getApplication: vi.fn(),
    addChangeListener: vi.fn(),
    removeChangeListener: vi.fn(),
  })),
}));

vi.mock('../../api/API', async () => ({
  ...(await vi.importActual('../../api/API')),
  getPackages: vi.fn().mockResolvedValue({ packages: [] }),
}));

import '../../i18n/config.ts';

import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { render, screen } from '@testing-library/react';

import List from '../../components/Packages/List.tsx';
import themes from '../../lib/themes';

const mockGetCachedApplication = vi.fn();
const mockGetApplication = vi.fn();
const mockAddChangeListener = vi.fn();
const mockRemoveChangeListener = vi.fn();

const mockStore = {
  getCachedApplication: mockGetCachedApplication,
  getApplication: mockGetApplication,
  addChangeListener: mockAddChangeListener,
  removeChangeListener: mockRemoveChangeListener,
};

describe('List Component', () => {
  const minProps = {
    appID: 'app123',
  };

  beforeEach(() => {
    mockGetCachedApplication.mockReset();
    mockGetApplication.mockReset();
    mockAddChangeListener.mockReset();
    mockRemoveChangeListener.mockReset();
  });

  it('should render the list', () => {
    mockStore.getCachedApplication.mockReturnValueOnce(null);

    render(
      <StyledEngineProvider injectFirst>
        <ThemeProvider theme={themes['light']}>
          <List {...minProps} />
        </ThemeProvider>
      </StyledEngineProvider>
    );
    expect(screen.getByText('Packages')).toBeTruthy();
  });
});
