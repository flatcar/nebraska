import { describe, expect, it } from 'vitest';

import { makeGroup } from '../../TestHelpers/groups';
import { effectivePolicy, isOverridden, isPausedHere } from '../../utils/groupPolicy';

describe('effectivePolicy', () => {
  it('takes the override where the node has one and the default elsewhere', () => {
    const group = makeGroup({ policy_updates_enabled: false, policy_max_updates_per_period: 9 });
    const effective = effectivePolicy(group);

    expect(effective.policy_updates_enabled).toBe(false);
    expect(effective.policy_max_updates_per_period).toBe(9);
    expect(effective.policy_period_interval).toBe('15 minutes');
    expect(group.policy_updates_enabled).toBe(true);
  });

  it('is the policy alone, which on a single node is the group policy', () => {
    const group = makeGroup(null);

    expect(effectivePolicy(group)).toEqual({
      policy_updates_enabled: true,
      policy_safe_mode: false,
      policy_office_hours: false,
      policy_timezone: 'Europe/Berlin',
      policy_period_interval: '15 minutes',
      policy_max_updates_per_period: 2,
      policy_update_timeout: '60 minutes',
    });
    expect(isOverridden(group, 'policy_updates_enabled')).toBe(false);
    expect(isPausedHere(group)).toBe(false);
  });

  it('tells which settings are overridden', () => {
    const group = makeGroup({ policy_timezone: 'Asia/Tokyo', policy_updates_enabled: false });

    expect(isOverridden(group, 'policy_office_hours', 'policy_timezone')).toBe(true);
    expect(isOverridden(group, 'policy_safe_mode')).toBe(false);
    expect(isPausedHere(group)).toBe(true);
  });
});
