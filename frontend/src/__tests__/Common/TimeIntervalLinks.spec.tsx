import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import TimeIntervalLinks from '../../components/common/TimeIntervalLinks';
import { defaultTimeInterval, timeIntervalsDefault } from '../../utils/helpers';

describe('TimeIntervalLinks', () => {
  it('should render correct time Interval links', () => {
    const { getByText } = render(
      <TimeIntervalLinks
        selectedInterval={defaultTimeInterval as unknown as string}
        intervalChangeHandler={() => {}}
      />
    );
    timeIntervalsDefault.forEach(timeInterval => {
      expect(getByText(timeInterval.displayValue)).toBeTruthy();
    });
  });
});
