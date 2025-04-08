import '../../i18n/config.ts';

import { fireEvent, render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import ConfirmationContent from '../../components/common/ConfirmationContent';
import { applicationsStore } from '../../stores/Stores';

const mockResolver = vi.fn();
describe('Confirmation Content', () => {
  const minProps = {
    data: {
      type: 'application',
      appID: '123',
    },
  };
  it('should render Confirmation Content correctly', () => {
    const { asFragment } = render(<ConfirmationContent data={minProps.data} />);
    expect(asFragment()).toMatchSnapshot();
  });
  it('should call delete handler function on yes confirmation', () => {
    applicationsStore().deleteApplication = mockResolver;
    const deleteApp = applicationsStore().deleteApplication;
    const { getByText } = render(<ConfirmationContent data={minProps.data} />);
    fireEvent.click(getByText('Yes'));
    expect(mockResolver).toHaveBeenCalled();
    applicationsStore().deleteApplication = deleteApp;
  });
});
