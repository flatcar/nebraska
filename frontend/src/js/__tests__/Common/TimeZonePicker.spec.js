import '../../i18n/config.ts';
import { fireEvent, render } from '@testing-library/react';
import React from 'react';
import TimezonePicker from '../../components/Common/TimezonePicker';

describe('TimeZonePicker', () => {
  it('should render suggestions on inputing timezone', () => {
    const { getByTestId, getByPlaceholderText, getByText } = render(<TimezonePicker />);
    fireEvent.click(getByTestId('timezone-readonly-input'));
    fireEvent.change(getByPlaceholderText('Start typing to search a timezone'), {
      target: { value: 'Asia/Calcutta' },
    });
    expect(getByText('Asia/Calcutta')).toBeInTheDocument();
  });
});
