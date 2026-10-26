import '../../i18n/config.ts';

import { act, fireEvent, render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

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

  it('should select a timezone tapped right after the highlight changes', async () => {
    const onSelect = vi.fn();
    const { getByTestId, getByPlaceholderText, getByText, getByRole } = render(
      <TimezonePicker value="Europe/Berlin" onSelect={onSelect} />
    );
    await act(async () => fireEvent.click(getByTestId('timezone-readonly-input')));
    fireEvent.change(getByPlaceholderText('Start typing to search a timezone'), {
      target: { value: 'Europe/Bu' },
    });
    const option = getByText('Europe/Bucharest').closest('[role="option"]')!;
    fireEvent.mouseMove(option);
    // The row must not be remounted, or the click lands on a detached node.
    expect(getByText('Europe/Bucharest').closest('[role="option"]')).toBe(option);
    fireEvent.click(option);
    fireEvent.click(getByRole('button', { name: 'Select' }));
    expect(onSelect).toHaveBeenCalledWith('Europe/Bucharest');
  });
});
