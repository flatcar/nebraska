import { Group, GroupLocalOverrides } from '../api/apiDataTypes';

const noOverrides: GroupLocalOverrides = {
  policy_updates_enabled: null,
  policy_safe_mode: null,
  policy_office_hours: null,
  policy_timezone: null,
  policy_period_interval: null,
  policy_max_updates_per_period: null,
  policy_update_timeout: null,
};

// A group as a control or edge node returns it: the policy fields hold the
// default, and local_overrides this node's overrides. Pass null for a group as
// a single node returns it.
export function makeGroup(overrides: Partial<GroupLocalOverrides> | null = {}): Group {
  return {
    id: 'group-1',
    name: 'Stable',
    description: '',
    created_ts: '2026-01-01T00:00:00Z',
    rollout_in_progress: false,
    application_id: 'app-1',
    channel_id: null,
    policy_updates_enabled: true,
    policy_safe_mode: false,
    policy_office_hours: false,
    policy_timezone: 'Europe/Berlin',
    policy_period_interval: '15 minutes',
    policy_max_updates_per_period: 2,
    policy_update_timeout: '60 minutes',
    channel: null as unknown as Group['channel'],
    track: 'stable',
    local_overrides: overrides === null ? null : { ...noOverrides, ...overrides },
  };
}
