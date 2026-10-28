import { Group, GroupLocalOverrides } from '../api/apiDataTypes';

export type PolicyField = keyof GroupLocalOverrides;

export type GroupPolicy = Pick<Group, PolicyField>;

// The policy in effect on this node: the node's override where it has one, and
// the group default elsewhere.
export function effectivePolicy(group: Group): GroupPolicy {
  const overrides = group.local_overrides;
  return {
    policy_updates_enabled: overrides?.policy_updates_enabled ?? group.policy_updates_enabled,
    policy_safe_mode: overrides?.policy_safe_mode ?? group.policy_safe_mode,
    policy_office_hours: overrides?.policy_office_hours ?? group.policy_office_hours,
    policy_timezone: overrides?.policy_timezone ?? group.policy_timezone,
    policy_period_interval: overrides?.policy_period_interval ?? group.policy_period_interval,
    policy_max_updates_per_period:
      overrides?.policy_max_updates_per_period ?? group.policy_max_updates_per_period,
    policy_update_timeout: overrides?.policy_update_timeout ?? group.policy_update_timeout,
  };
}

export function isOverridden(group: Group, ...fields: PolicyField[]) {
  return fields.some(
    field => group.local_overrides?.[field] !== null && group.local_overrides?.[field] !== undefined
  );
}

export function isPausedHere(group: Group) {
  return group.local_overrides?.policy_updates_enabled === false;
}
