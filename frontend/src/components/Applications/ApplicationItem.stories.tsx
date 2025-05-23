import { Meta, StoryFn } from '@storybook/react';
import { JSX } from 'react/jsx-runtime';
import { MemoryRouter } from 'react-router';

import API, { APIContext } from '../../api/API';
import ApplicationItem, { ApplicationItemProps } from './ApplicationItem';

export default {
  title: 'applications/ApplicationItem',
  argTypes: {
    onUpdate: { action: 'onUpdate' },
  },
} as Meta;

const Template: StoryFn<ApplicationItemProps> = (
  args: JSX.IntrinsicAttributes & ApplicationItemProps
) => {
  class APIMock extends API {
    static getInstancesCount(): Promise<number> {
      return new Promise(resolve => resolve(20));
    }
  }

  return (
    <APIContext.Provider value={APIMock}>
      <MemoryRouter>
        <ApplicationItem {...args} />
      </MemoryRouter>
    </APIContext.Provider>
  );
};

export const NoGroups = {
  render: Template,

  args: {
    numberOfInstances: 1,
    id: '123',
    productId: 'xxx-xxx-yyy',
    name: 'ABC',
    description: 'App Item Description',
    groups: [],
  },
};

export const Application = {
  render: Template,

  args: {
    numberOfInstances: 1,
    id: '123',
    productId: 'xxx-xxx-yyy',
    name: 'ABC',
    description: 'App Item Description',
    groups: [
      {
        id: 'xxx-xxx-1',
        name: 'First group',
        description: '',
        created_ts: '',
        rollout_in_progress: false,
        application_id: '',
        channel_id: null,
        policy_updates_enabled: false,
        policy_safe_mode: false,
        policy_office_hours: false,
        policy_timezone: null,
        policy_period_interval: '',
        policy_max_updates_per_period: 0,
        policy_update_timeout: '',

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
            extra_files: [],
          },
          arch: 0,
        },
        track: '',
      },
    ],
  },
};
