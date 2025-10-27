import { getInstanceStatus } from '../utils/helpers';

type Arch = number;

export interface Group {
  id: string;
  name: string;
  description: string;
  created_ts: string;
  rollout_in_progress: boolean;
  application_id: string;
  channel_id: null | string;
  policy_updates_enabled: boolean;
  policy_safe_mode: boolean;
  policy_office_hours: boolean;
  policy_timezone: null | string;
  policy_period_interval: string;
  policy_max_updates_per_period: number;
  policy_update_timeout: string;
  channel: Channel;
  track: string;
}

export interface Channel {
  id: string;
  name: string;
  color: string;
  created_ts: string;
  application_id: string;
  package_id: null | string;
  package: null | Package;
  arch: Arch;
}

export interface File {
  id?: number;
  name: string;
  hash: string;
  hash256: string;
  size: string;
  created_ts?: string;
}

export interface Package {
  id?: string;
  type: number;
  version: string;
  url: string;
  filename: null | string;
  description: null | string;
  size: null | string;
  hash: null | string;
  created_ts: string;
  channels_blacklist: null | string[];
  application_id: string;
  flatcar_action?: null | FlatcarAction;
  arch: Arch;
  extra_files: File[];
  is_floor?: boolean;
  floor_reason?: string | null;
}

export interface FlatcarAction {
  id?: string;
  event?: string;
  chromeos_version?: string;
  sha256?: string;
  needs_admin?: boolean;
  is_delta?: boolean;
  disable_payload_backoff?: boolean;
  metadata_signature_rsa?: string;
  metadata_size?: string;
  deadline?: string;
  created_ts?: string;
  package_id?: string;
}

export type Packages = { totalCount: number; items: Package[] };

export interface Application {
  id: string;
  name: string;
  description: string;
  product_id: string;
  created_ts: string;
  team_id: string;
  groups: Group[];
  channels: Channel[];
  packages: Packages | null | undefined;
  instances: {
    count: number;
  };
}

export interface Activity {
  id: number;
  app_id: string;
  group_id: string;
  created_ts: string;
  class: number;
  severity: number;
  version: string;
  application_name: string;
  group_name: string | null;
  channel_name: string | null;
  instance_id: string | null;
}

export interface Instance {
  id: string;
  alias: string;
  created_ts: string | Date | number;
  ip: string;
  application: InstanceApplication;
  statusInfo?: ReturnType<typeof getInstanceStatus>;
  statusHistory?: InstanceStatusHistory[];
}

export interface Instances {
  total: number;
  instances: Instance[];
}

export interface InstanceApplication {
  instance_id: string;
  application_id: string;
  group_id: string;
  version: string;
  created_ts: string | Date | number;
  status: null | number;
  last_check_for_updates: string;
}

export interface InstanceStatusHistory {
  status: number;
  version: string;
  created_ts: string | Date | number;
  error_code: number | null;
}

export interface VersionBreakdownEntry {
  version: string;
  instances: number;
  percentage: number;
}
