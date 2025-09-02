import '../../i18n/config.ts';

import { act, fireEvent, render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import TimezonePicker from '../../components/common/TimezonePicker';

describe('TimeZonePicker', () => {
  it('should render suggestions on inputting timezone', async () => {
    const { getByTestId, getByPlaceholderText, getByText } = render(
      <TimezonePicker value="" onSelect={() => {}} />
    );
    await act(async () => fireEvent.click(getByTestId('timezone-readonly-input')));
    fireEvent.change(getByPlaceholderText('Start typing to search a timezone'), {
      target: { value: 'Asia/Calcutta' },
    });
    expect(getByText('Asia/Calcutta')).toBeTruthy();
  });
});
