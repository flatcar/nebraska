import '../../i18n/config.ts';
import { fireEvent, render } from '@testing-library/react';
import React from 'react';
import ConfirmationContent from '../../components/common/ConfirmationContent';
import { applicationsStore } from '../../stores/Stores';

function mockResolver() {
  return Promise.resolve([]);
}
mockResolver = vi.fn();
describe('Confirmation Content', () => {
  const minProps = {
    data: {
      type: 'application',
      appID: '123',
    },
  };
  it('should render Confirmation Content correctly', () => {
    const { asFragment } = render(<ConfirmationContent data={minProps.data} channel={{}} />);
    expect(asFragment()).toMatchSnapshot();
  });
  it('should call delete handler function on yes confirmation', () => {
    applicationsStore().deleteApplication = mockResolver;
    const deleteApp = applicationsStore().deleteApplication;
    const { getByText } = render(<ConfirmationContent data={minProps.data} channel={{}} />);
    fireEvent.click(getByText('Yes'));
    expect(mockResolver).toHaveBeenCalled();
    applicationsStore().deleteApplication = deleteApp;
  });
});
