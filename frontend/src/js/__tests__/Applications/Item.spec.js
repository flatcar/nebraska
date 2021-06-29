import '../../i18n/config.ts';
import { render } from '@testing-library/react';
import React from 'react';
import { BrowserRouter } from 'react-router-dom';
import Item from '../../components/Applications/Item';

describe('Application Item', () => {
  it('should render item correctly for no data', () => {
    const { asFragment } = render(
      <BrowserRouter>
        <Item
          application={{
            instances: {},
            id: '123',
            name: 'ABC',
          }}
          handleUpdateApplication={() => {}}
        />
      </BrowserRouter>
    );
    expect(asFragment()).toMatchSnapshot();
  });
  it('should render item correctly for data', () => {
    const minProps = {
      instances: {
        count: '1',
      },
      id: '123',
      name: 'ABC',
      group: {
        channel: {
          id: 'DEF',
          name: 'main',
          color: '#777777',
          created_ts: '2018-10-16T21:07:56.819939+05:30',
          application_id: '123',
          package_id: 'XYZ',
          package: {
            id: 'PACK_ID',
            type: 4,
            version: '1.11.3',
            url: 'https://github.com/kinvolk',
            filename: '',
            description: '',
            size: '',
            hash: '',
            created_ts: '2019-07-18T20:10:39.163326+05:30',
            channels_blacklist: null,
            application_id: 'df1c8bbb-f525-49ee-8c94-3ca548b42059',
            flatcar_action: null,
            arch: 0,
          },
          arch: 0,
        },
      },
      description: 'App Item Description',
    };
    const { getByText } = render(
      <BrowserRouter>
        <Item application={{ ...minProps }} handleUpdateApplication={() => {}} />
      </BrowserRouter>
    );
    expect(getByText(minProps.description)).toBeInTheDocument();
    expect(getByText(minProps.instances.count)).toBeInTheDocument();
    expect(getByText(minProps.name)).toBeInTheDocument();
    expect(getByText(minProps.id)).toBeInTheDocument();
  });
});
