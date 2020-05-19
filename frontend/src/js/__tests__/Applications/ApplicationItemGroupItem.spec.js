import {getByText, render, waitForDomChange} from '@testing-library/react';
import React from 'react';
import { BrowserRouter } from 'react-router-dom';
import API from '../../api/API';
import ApplicationItemGroupItem from '../../components/Applications/ApplicationItemGroupItem.react';

function mockResolver() {
  return Promise.resolve(1);
}
const mockAjax = jest.fn(() => mockResolver);
describe('Application Item Group Item', () => {
  const minProps = {
    group: {
      name: 'ABC',
      application_id: '123',
      id: '1'
    },
    appName: 'FlatCar'
  };
  beforeEach(() => {
    API.getInstancesCount = mockAjax();
  });
  it('should render correct link and correct total instances', async () => {
    const {container, getByText} = render(
      <BrowserRouter>
        <ApplicationItemGroupItem {...minProps}/>
      </BrowserRouter>);
    await waitForDomChange(container);
    expect(container.querySelector('a').getAttribute('href'))
      .toBe(`/apps/${minProps.group.application_id}/groups/${minProps.group.id}`);
    expect(getByText(`${minProps.group.name} (1)`)).toBeInTheDocument();
  });
  afterEach(() => {
    mockAjax.mockClear();
  });
});
