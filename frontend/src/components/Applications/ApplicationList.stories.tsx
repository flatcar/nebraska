import { Meta, StoryFn } from '@storybook/react';
import { MemoryRouter } from 'react-router-dom';
import API, { APIContext } from '../../api/API';
import { ApplicationListPure, ApplicationListPureProps } from './ApplicationList';

export default {
  title: 'applications/ApplicationList',
} as Meta;

const Template: StoryFn<ApplicationListPureProps> = args => {
  class APIMock extends API {
    /* eslint-disable no-unused-vars */
    static getInstancesCount(
      applicationID: string,
      groupID: string,
      duration: string,
    ): Promise<number> {
      return new Promise(resolve => resolve(20));
    }
  }

  return (
    <APIContext.Provider value={APIMock}>
      <MemoryRouter>
        <ApplicationListPure {...args} />
      </MemoryRouter>
    </APIContext.Provider>
  );
};

export const Loading = {
  render: Template,

  args: {
    applications: null,
    loading: true,
  },
};

export const Applications = {
  render: Template,

  args: {
    applications: [
      {
        id: 'FFFF-AAAA-CCCC-EEEE',
        product_id: 'xxx-xxx-yyy',
        name: 'ABC',
        description: 'App Item Description',
        created_ts: '2018-10-16T21:07:56.819939+05:30',
        team_id: 'YYY-XXX-xxx',
        channels: [],
        packages: [],
        instances: { count: 20 },
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
      {
        id: 'BBBB-AAAA-CCCC-EEEE',
        product_id: 'xxx-xxx-yyy',
        name: 'ABC 2 the return',
        description: 'App Item Description',
        created_ts: '2021-10-17T21:07:56.819939+05:30',
        team_id: 'YYY-XXX-xxx',
        channels: [],
        packages: [],
        instances: { count: 20 },
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
    ],
    loading: false,
  },
};

export const EditOpen = {
  render: Template,

  args: {
    ...Applications.args,
    editOpen: true,
    editId: 'BBBB-AAAA-CCCC-EEEE',
  },
};

export const SearchTerm = {
  render: Template,

  args: {
    ...Applications.args,
    defaultSearchTerm: 'the return',
  },
};
