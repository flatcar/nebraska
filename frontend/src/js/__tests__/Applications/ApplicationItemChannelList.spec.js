import {render} from '@testing-library/react';
import React from 'react';
import ApplicationItemChannelsList from '../../components/Applications/ApplicationItemChannelsList.react';

describe('Application Channel List ', () => {
  it('should render no application channel for empty channels', () => {
    const {container} = render(<ApplicationItemChannelsList channels={[]}/>);
    expect(container.firstChild).toBeEmpty();
  });
  it('should render channel item if there are channels', () => {
    const channels = [{
      'id': 'ABC',
      'name': 'main',
      'color': '#777777',
      'created_ts': '2018-10-16T21:07:56.819939+05:30',
      'application_id': 'DEF',
      'package_id': 'XYZ',
      'package': {'id': 'PACK_ID', 'type': 4,
                  'version': '1.11.3', 'url': 'https://github.com/kinvolk',
                  'filename': '', 'description': '', 'size': '', 'hash': '', 'created_ts': '2019-07-18T20:10:39.163326+05:30',
                  'channels_blacklist': null, 'application_id': 'df1c8bbb-f525-49ee-8c94-3ca548b42059', 'flatcar_action': null, 'arch': 0},
      'arch': 0}];
    const {asFragment} = render(<ApplicationItemChannelsList channels={channels}/>);
    expect(asFragment()).toMatchSnapshot();
  });
});
