import '../../i18n/config.ts';
import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { render, screen } from '@testing-library/react';
import API from '../../api/API';
import List from '../../components/Packages/List.tsx';
import themes from '../../lib/themes';
import { applicationsStore } from '../../stores/Stores';

jest.mock('../../api/API');
jest.mock('../../stores/Stores', () => ({
  applicationsStore: jest.fn(),
}));

describe('List Component', () => {
  const minProps = {
    appID: 'app123',
  };

  const mockGetCachedApplication = jest.fn();
  const mockGetApplication = jest.fn();
  const mockAddChangeListener = jest.fn();
  const mockRemoveChangeListener = jest.fn();

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
    expect(screen.getByText('Packages')).toBeInTheDocument();
  });
});
