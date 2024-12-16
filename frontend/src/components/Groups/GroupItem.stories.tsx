import { Meta, Story } from '@storybook/react';
import { MemoryRouter } from 'react-router-dom';
import { PureGroupItem, PureGroupItemProps } from './GroupItem';

export default {
  title: 'groups/GroupItem',
  argTypes: {
    handleUpdateGroup: { action: 'handleUpdateGroup' },
    deleteGroup: { action: 'deleteGroup' },
  },
} as Meta;

const Template: Story<PureGroupItemProps> = args => {
  return (
    <MemoryRouter>
      <PureGroupItem {...args} />
    </MemoryRouter>
  );
};

export const Group = Template.bind({});

Group.args = {
  versionBreakdown: [],
  totalInstances: 2,
  group: {
    id: '11a585f6-9418-4df0-8863-78b2fd3240f8',
    name: 'Stable (ARM)',
    description: 'For production clusters (ARM)',
    created_ts: '2015-09-19T07:09:34.269062+02:00',
    rollout_in_progress: false,
    application_id: 'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
    channel_id: '5dfe7b12-c94a-470d-a2b6-2eae78c5c9f5',
    policy_updates_enabled: true,
    policy_safe_mode: false,
    policy_office_hours: false,
    policy_timezone: 'Europe/Berlin',
    policy_period_interval: '1 minutes',
    policy_max_updates_per_period: 999999,
    policy_update_timeout: '60 minutes',
    channel: {
      id: '5dfe7b12-c94a-470d-a2b6-2eae78c5c9f5',
      name: 'stable',
      color: '#1458d6',
      created_ts: '2015-09-19T07:09:34.261241+02:00',
      application_id: 'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
      package_id: null,
      package: null,
      arch: 2,
    },
    track: 'stable',
  },
};

export const Loading = Template.bind({});

Loading.args = {
  versionBreakdown: null,
  totalInstances: null,
  group: {
    id: '11a585f6-9418-4df0-8863-78b2fd3240f8',
    name: 'Stable (ARM)',
    description: 'For production clusters (ARM)',
    created_ts: '2015-09-19T07:09:34.269062+02:00',
    rollout_in_progress: false,
    application_id: 'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
    channel_id: '5dfe7b12-c94a-470d-a2b6-2eae78c5c9f5',
    policy_updates_enabled: true,
    policy_safe_mode: false,
    policy_office_hours: false,
    policy_timezone: 'Europe/Berlin',
    policy_period_interval: '1 minutes',
    policy_max_updates_per_period: 999999,
    policy_update_timeout: '60 minutes',
    channel: {
      id: '5dfe7b12-c94a-470d-a2b6-2eae78c5c9f5',
      name: 'stable',
      color: '#1458d6',
      created_ts: '2015-09-19T07:09:34.261241+02:00',
      application_id: 'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
      package_id: null,
      package: null,
      arch: 2,
    },
    track: 'stable',
  },
};
