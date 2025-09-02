import '../../i18n/config.ts';

import { act, fireEvent, render } from '@testing-library/react';
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
  it('should call delete handler function on yes confirmation', async () => {
    applicationsStore().deleteApplication = mockResolver;
    const deleteApp = applicationsStore().deleteApplication;
    const { getByText } = render(<ConfirmationContent data={minProps.data} />);
    await act(async () => fireEvent.click(getByText('Yes')));
    expect(mockResolver).toHaveBeenCalled();
    applicationsStore().deleteApplication = deleteApp;
  });
});
