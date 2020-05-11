import {fireEvent, render} from '@testing-library/react';
import React from 'react';
import ConfirmationContent from '../../components/Common/ConfirmationContent.react';
import ApplicationsStore from '../../stores/ApplicationsStore';

function mockResolver() {
  return Promise.resolve([]);
}
mockResolver = jest.fn();
describe('Confirmation Content', () => {
  const minProps = {
    data: {
      type: 'application',
      appID: '123'
    }
  };
  it('should render Confirmation Content correctly', () => {
    const {asFragment} = render(
      <ConfirmationContent
        data={minProps.data}
        channel={{}}
      />);
    expect(asFragment()).toMatchSnapshot();
  });
  it('should call delete handler function on yes confirmation', () => {
    ApplicationsStore.deleteApplication = mockResolver;
    const deleteApp = ApplicationsStore.deleteApplication;
    const {getByText} = render(
      <ConfirmationContent
        data={minProps.data}
        channel = {{}}
      />);
    fireEvent.click(getByText('Yes'));
    expect(mockResolver).toHaveBeenCalled();
    ApplicationsStore.deleteApplication = deleteApp;
  });
});
