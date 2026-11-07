import { Meta, StoryFn } from '@storybook/react-vite';

import { Group } from '../../api/apiDataTypes';
import GroupPolicyPanes, { GroupPolicyPanesProps } from './GroupPolicyPanes';

export default {
  title: 'groups/GroupPolicyPanes',
  argTypes: {
    onEditDefault: { action: 'onEditDefault' },
    onEditOverrides: { action: 'onEditOverrides' },
    onPauseUpdates: { action: 'onPauseUpdates' },
    onResumeUpdates: { action: 'onResumeUpdates' },
  },
} as Meta;

const Template: StoryFn<GroupPolicyPanesProps> = args => <GroupPolicyPanes {...args} />;

// On a control or edge node the policy fields hold the group default.
const group: Group = {
  id: '11a585f6-9418-4df0-8863-78b2fd3240f8',
  name: 'Stable (AMD64)',
  description: 'For production clusters (AMD64)',
  created_ts: '2015-09-19T07:09:34.269062+02:00',
  rollout_in_progress: false,
  application_id: 'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
  channel_id: null,
  policy_updates_enabled: true,
  policy_safe_mode: true,
  policy_office_hours: false,
  policy_timezone: 'Europe/Berlin',
  policy_period_interval: '15 minutes',
  policy_max_updates_per_period: 2,
  policy_update_timeout: '60 minutes',
  channel: null as unknown as Group['channel'],
  track: 'stable',
  local_overrides: {
    policy_updates_enabled: null,
    policy_safe_mode: null,
    policy_office_hours: null,
    policy_timezone: null,
    policy_period_interval: null,
    policy_max_updates_per_period: null,
    policy_update_timeout: null,
  },
};

export const ControlNode = {
  render: Template,
  args: { group, instanceMode: 'control' },
};

// Safe mode paused this edge, which also runs its own rollout rate and hours.
export const EdgeNodeWithOverrides = {
  render: Template,
  args: {
    instanceMode: 'edge',
    group: {
      ...group,
      local_overrides: {
        ...group.local_overrides,
        policy_updates_enabled: false,
        policy_office_hours: true,
        policy_timezone: 'Asia/Tokyo',
        policy_max_updates_per_period: 5,
        policy_period_interval: '2 hours',
      },
    },
  },
};
