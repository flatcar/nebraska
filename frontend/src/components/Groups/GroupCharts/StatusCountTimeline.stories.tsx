import { Meta, StoryFn } from '@storybook/react';
import { JSX } from 'react/jsx-runtime';
import { MemoryRouter } from 'react-router-dom';

import GroupChartsStore from '../../../stores/GroupChartsStore';
import { groupChartStoreContext } from '../../../stores/Stores';
import StatusCountTimeline, { StatusCountTimelineProps } from './StatusCountTimeline';

export default {
  title: 'groups/StatusCountTimeline',
} as Meta;

const statusTimelineData = {
  '2021-11-07T10:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-07T11:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-07T12:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-07T13:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-07T14:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-07T15:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-07T16:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-07T17:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-07T18:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-07T19:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-07T20:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-07T21:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-07T22:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-07T23:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-08T00:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-08T01:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-08T02:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-08T03:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-08T04:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-08T05:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-08T06:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-08T07:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-08T08:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
  '2021-11-08T09:35:28.823161+01:00': {
    '1': { '2191.5.0': 177 },
    '2': { '2191.5.0': 221 },
    '3': { '2191.5.0': 12 },
    '6': { '2191.5.0': 193 },
    '7': { '2191.5.0': 215 },
  },
  '2021-11-08T10:35:28.823161+01:00': { '1': {}, '2': {}, '3': {}, '6': {}, '7': {} },
};

const Template: StoryFn<StatusCountTimelineProps> = (
  args: JSX.IntrinsicAttributes & StatusCountTimelineProps
) => {
  class GroupChartsStoreMock extends GroupChartsStore {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    async getGroupStatusCountTimeline(appID: string, groupID: string, duration: string) {
      return statusTimelineData;
    }
  }

  const ChartStoreContext = groupChartStoreContext();

  return (
    <ChartStoreContext.Provider value={new GroupChartsStoreMock()}>
      <MemoryRouter>
        <StatusCountTimeline {...args} />
      </MemoryRouter>
    </ChartStoreContext.Provider>
  );
};

export const Timeline = {
  render: Template,

  args: {
    isAnimationActive: false,
    group: {
      id: '9a2deb70-37be-4026-853f-bfdd6b347bbe',
      name: 'Stable (AMD64)',
      description: 'For production clusters (AMD64)',
      created_ts: '2015-09-19T07:09:34.269062+02:00',
      rollout_in_progress: true,
      application_id: 'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
      channel_id: 'e06064ad-4414-4904-9a6e-fd465593d1b2',
      policy_updates_enabled: true,
      policy_safe_mode: false,
      policy_office_hours: false,
      policy_timezone: 'Europe/Berlin',
      policy_period_interval: '1 minutes',
      policy_max_updates_per_period: 999999,
      policy_update_timeout: '60 minutes',
      channel: {
        id: 'e06064ad-4414-4904-9a6e-fd465593d1b2',
        name: 'stable',
        color: '#14b9d6',
        created_ts: '2015-09-19T07:09:34.261241+02:00',
        application_id: 'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
        package_id: '84b4c599-9b6b-44a8-b13c-d4263fff0403',
        package: {
          id: '84b4c599-9b6b-44a8-b13c-d4263fff0403',
          type: 1,
          version: '2191.5.0',
          url: 'https://update.release.flatcar-linux.net/amd64-usr/2191.5.0/',
          filename: 'flatcar_production_update.gz',
          description: 'Flatcar Container Linux 2191.5.0',
          size: '465881871',
          hash: 'r3nufcxgMTZaxYEqL+x2zIoeClk=',
          created_ts: '2019-09-05T12:41:09.265687+02:00',
          channels_blacklist: null,
          application_id: 'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
          flatcar_action: {
            id: '1f6e1bcf-4ebb-4fe6-8ca3-2cb6ad90d5dd',
            event: 'postinstall',
            chromeos_version: '',
            sha256: 'LIkAKVZY2EJFiwTmltiJZLFLA5xT/FodbjVgqkyF/y8=',
            needs_admin: false,
            is_delta: false,
            disable_payload_backoff: true,
            metadata_signature_rsa: '',
            metadata_size: '',
            deadline: '',
            created_ts: '2019-08-20T02:12:37.532281+02:00',
          },
          arch: 1,
          extra_files: [],
        },
        arch: 1,
      },
      track: 'stable',
    },
    duration: { displayValue: '1 day', queryValue: '1d', disabled: false },
  },
};
