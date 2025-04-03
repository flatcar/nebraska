import '../../i18n/config.ts';

import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import API from '../../api/API';
import List from '../../components/Packages/List.tsx';
import themes from '../../lib/themes';
import { applicationsStore } from '../../stores/Stores';

vi.mock('../../api/API');
vi.mock('../../stores/Stores', () => ({
  applicationsStore: vi.fn(),
}));

describe('List Component', () => {
  const minProps = {
    appID: 'app123',
  };

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

  beforeEach(() => {
    applicationsStore.mockReturnValue(mockStore);
    API.getPackages.mockResolvedValueOnce({ packages: [] });

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
