import { render } from '@testing-library/react';
import React from 'react';
import TimeIntervalLinks from '../../components/Common/TimeIntervalLinks';
import { defaultTimeInterval, timeIntervalsDefault } from '../../utils/helpers';

describe('TimeIntervalLinks', () => {
  it('should render correct time Interval links', () => {
    const { getByText } = render(<TimeIntervalLinks selectedInterval={defaultTimeInterval} />);
    timeIntervalsDefault.forEach(timeInterval => {
      expect(getByText(timeInterval.displayValue)).toBeInTheDocument();
    });
  });
});
