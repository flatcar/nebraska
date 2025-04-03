import '../../i18n/config.ts';
import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { render } from '@testing-library/react';
import SearchInput from '../../components/common/ListSearch';
import themes from '../../lib/themes';

describe('List Search', () => {
  it('should render correct ListSearch', () => {
    const { asFragment } = render(
      <StyledEngineProvider injectFirst>
        <ThemeProvider theme={themes['light']}>
          <SearchInput />
        </ThemeProvider>
      </StyledEngineProvider>
    );
    expect(asFragment()).toMatchSnapshot();
  });
});
