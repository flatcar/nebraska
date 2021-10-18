import { MuiThemeProvider } from '@material-ui/core/styles';
import { render, screen } from '@testing-library/react';
import jest from 'jest-mock';
import React from 'react';
import { BrowserRouter } from 'react-router-dom';
import API from '../../api/API';
import ApplicationItemGroupItem from '../../components/Applications/ApplicationItemGroupItem';
import { theme } from '../../TestHelpers/theme';

function mockResolver() {
  return Promise.resolve(1);
}
const mockAjax = jest.fn(() => mockResolver);
describe('Application Item Group Item', () => {
  const minProps = {
    group: {
      name: 'ABC',
      application_id: '123',
      id: '1',
      channel: {
        name: 'main',
      },
    },
    appName: 'FlatCar',
  };
  beforeEach(() => {
    API.getInstancesCount = mockAjax();
  });
  it('should render correct link and correct total instances', async () => {
    render(
      <BrowserRouter>
        <MuiThemeProvider theme={theme}>
          <ApplicationItemGroupItem {...minProps} />
        </MuiThemeProvider>
      </BrowserRouter>
    );

    expect(screen.getByText(`${minProps.group.name}`)).toBeInTheDocument();
    expect(screen.getByText(minProps.group.channel.name)).toBeInTheDocument();
    expect(
      screen.getByRole('link', {
        href: `/apps/${minProps.group.application_id}/groups/${minProps.group.id}`,
      })
    ).toBeInTheDocument();
  });
  afterEach(() => {
    mockAjax.mockClear();
  });
});
