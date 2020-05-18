import {render, waitForDomChange} from '@testing-library/react';
import React from 'react';
import { BrowserRouter } from 'react-router-dom';
import List from '../../components/Applications/List.react';
import { applicationsStore } from '../../stores/Stores';

describe('Application List', () => {
  const minProps = {
    'applications': [{
      'id': '123',
      'name': 'ABC',
      'channels': [{
        'id': 'DEF',
        'name': 'main',
        'color': '#777777',
        'created_ts': '2018-10-16T21:07:56.819939+05:30',
        'application_id': '123',
        'package_id': 'XYZ',
        'package': {'id': 'PACK_ID', 'type': 4,
                    'version': '1.11.3', 'url': 'https://github.com/kinvolk',
                    'filename': '', 'description': '', 'size': '', 'hash': '', 'created_ts': '2019-07-18T20:10:39.163326+05:30',
                    'channels_blacklist': null, 'application_id': 'df1c8bbb-f525-49ee-8c94-3ca548b42059', 'flatcar_action': null, 'arch': 0},
        'arch': 0}],
      'description': 'App Item Description',
      'instances': {
        'count': '1'
      },
    }]

  };
  it('should render list correctly with data', () => {
    const getCachedApplications = applicationsStore.getCachedApplications();
    applicationsStore.getCachedApplications = () => minProps.applications;
    const {asFragment} = render(<BrowserRouter><List/></BrowserRouter>);
    expect(asFragment()).toMatchSnapshot();
    applicationsStore.getCachedApplications = getCachedApplications;
  });
});
