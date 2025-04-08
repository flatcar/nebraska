import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import TimeIntervalLinks from '../../components/common/TimeIntervalLinks';
import { defaultTimeInterval, timeIntervalsDefault } from '../../utils/helpers';

describe('TimeIntervalLinks', () => {
  it('should render correct time Interval links', () => {
    const appID = 'yourAppID';
    const groupID = 'yourGroupID';
    const intervalChangeHandler = () => { };
    const { getByText } = render(
      <TimeIntervalLinks
        selectedInterval={defaultTimeInterval.toString()}
        intervalChangeHandler={intervalChangeHandler}
        appID={appID}
        groupID={groupID}
      />);
    timeIntervalsDefault.forEach(timeInterval => {
      expect(getByText(timeInterval.displayValue)).toBeTruthy();
    });
  });
});
