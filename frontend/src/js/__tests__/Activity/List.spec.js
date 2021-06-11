import { render } from '@testing-library/react';
import React from 'react';
import { BrowserRouter } from 'react-router-dom';
import List from '../../components/Activity/List';
import { makeLocaleTime } from '../../utils/helpers';

describe('Activity List', () => {
  it('should render correct entries and timestamp for Activity list', () => {
    const minProps = {
      timestamp: 'Wed, 13 May 2020 14:56:03 GMT',
      entries: [
        {
          application_name: 'ABC',
          channel_name: 'beta',
          class: 6,
          created_ts: '2020-05-13T20:26:03.837688+05:30',
          group_name: null,
          instance_id: null,
          severity: 2,
          version: '0.0.0',
        },
        {
          application_name: 'DEF',
          channel_name: 'beta',
          class: 6,
          created_ts: '2020-05-13T20:25:52.589886+05:30',
          group_name: null,
          instance_id: null,
          severity: 2,
          version: '0.0.0',
        },
      ],
    };
    const time = makeLocaleTime(minProps.timestamp, {
      showTime: false,
      dateFormat: { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' },
    });
    const { getByText } = render(
      <BrowserRouter>
        <List {...minProps} />
      </BrowserRouter>
    );
    expect(getByText(time)).toBeInTheDocument();
    minProps.entries.forEach(entry => {
      expect(getByText(entry.application_name)).toBeInTheDocument();
    });
  });
});
