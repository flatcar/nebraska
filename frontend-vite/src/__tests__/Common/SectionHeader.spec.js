import '../../i18n/config.ts';
import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { render } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import SectionHeader from '../../components/common/SectionHeader';
import themes from '../../lib/themes';

describe('Section Header', () => {
  const minProps = {
    breadcrumbs: [
      {
        path: '/apps',
        label: 'Applications',
      },
    ],
    title: 'Flatcar',
  };
  it('renders section header correctly', () => {
    const { asFragment } = render(
      <StyledEngineProvider injectFirst>
        <ThemeProvider theme={themes['light']}>
          <BrowserRouter>
            <SectionHeader {...minProps} />
          </BrowserRouter>
        </ThemeProvider>
      </StyledEngineProvider>
    );
    expect(asFragment()).toMatchSnapshot();
  });
});
