import {render} from '@testing-library/react';
import React from 'react';
import Container from '../../components/Activity/Container.react';
import { activityStore } from '../../stores/Stores';

describe('Activity Container', () => {
  it('should render fetched activities correctly', () => {
    activityStore.activity = [{
      created_ts: '2020-05-13T20:26:03.837688+05:30',
      class: 6,
      severity: 2,
      version: '0.0.0',
      application_name: 'ABC',
      group_name: null,
      channel_name: 'beta',
      instance_id: null
    }, {
      created_ts: '2020-05-13T20:25:52.589886+05:30',
      class: 6,
      severity: 2,
      version: '0.0.0',
      application_name: 'DBZ',
      group_name: null,
      channel_name: 'beta',
      instance_id: null
    }];

    const {getByText} = render(<Container/>);
    expect(getByText('ABC')).toBeInTheDocument();
    expect(getByText('DBZ')).toBeInTheDocument();
    activityStore.activity = null;
  });
  it('should render no activity found message when there is no activity fetched', () => {
    activityStore.activity = [];
    const {getAllByTestId} = render(<Container/>);
    expect(getAllByTestId('empty')).toBeTruthy();
    activityStore.activity = null;
  });
});
