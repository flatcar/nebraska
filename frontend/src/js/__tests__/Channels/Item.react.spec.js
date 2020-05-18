import {fireEvent, render} from '@testing-library/react';
import React from 'react';
import Item from '../../components/Channels/Item.react';

describe('Channel Item', () => {
  const minProps = {
    channel: {
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
      'arch': 0},
    handleUpdateChannel: jest.fn(() => {})
  };
  it('should show confirmation box on delete click', () => {
    const {getByText} = render(
      <Item
        {...minProps}
      />);
    window.confirm = jest.fn(() => true);
    fireEvent.click(getByText('Delete'));
    expect(window.confirm).toBeCalled();
  });
  it('should call edit handler on edit click', () => {
    const {getByText} = render(<Item {...minProps}/>);
    fireEvent.click(getByText('Edit'));
    expect(minProps.handleUpdateChannel).toBeCalled();
  });
});
