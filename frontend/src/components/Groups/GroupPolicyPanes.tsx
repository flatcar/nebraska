import EditIcon from '@mui/icons-material/Edit';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import { Box, Button, Grid, Paper, Tooltip, Typography } from '@mui/material';
import { TFunction } from 'i18next';
import React from 'react';
import { useTranslation } from 'react-i18next';

import { Group } from '../../api/apiDataTypes';
import {
  effectivePolicy,
  GroupPolicy,
  isOverridden,
  isPausedHere,
  PolicyField,
} from '../../utils/groupPolicy';
import { InstanceMode } from '../../utils/instanceMode';
import { CardFeatureLabel, CardLabel } from '../common/Card';
import { formatUpdateLimits } from './GroupItem';
import OverrideLabel from './OverrideLabel';
import UpdatesStatus from './UpdatesStatus';

interface PolicyRow {
  label: string;
  fields: PolicyField[];
  value: (policy: GroupPolicy, pausedHere: boolean) => React.ReactNode;
}

function policyRows(t: TFunction): PolicyRow[] {
  return [
    {
      label: t('groups|updates'),
      fields: ['policy_updates_enabled'],
      value: (policy, pausedHere) => (
        <Box display="flex" alignItems="center" component="span">
          <UpdatesStatus enabled={policy.policy_updates_enabled} pausedHere={pausedHere} />
        </Box>
      ),
    },
    {
      label: t('groups|office_hours_only'),
      fields: ['policy_office_hours', 'policy_timezone'],
      value: policy =>
        policy.policy_office_hours
          ? t('groups|office_hours_in', { timezone: policy.policy_timezone })
          : t('frequent|no'),
    },
    {
      label: t('groups|safe_mode'),
      fields: ['policy_safe_mode'],
      value: policy => (policy.policy_safe_mode ? t('frequent|yes') : t('frequent|no')),
    },
    {
      label: t('groups|updates_policy'),
      fields: ['policy_max_updates_per_period', 'policy_period_interval'],
      value: policy => formatUpdateLimits(t, policy),
    },
    {
      label: t('groups|updates_timeout'),
      fields: ['policy_update_timeout'],
      value: policy => policy.policy_update_timeout,
    },
  ];
}

interface PolicyPaneProps {
  title: string;
  caption: string;
  action: React.ReactNode;
  testId: string;
  children: React.ReactNode;
}

function PolicyPane({ title, caption, action, testId, children }: PolicyPaneProps) {
  return (
    <Paper
      variant="outlined"
      component="section"
      aria-label={title}
      data-testid={testId}
      sx={{ height: '100%', p: 2 }}
    >
      <Box display="flex" alignItems="flex-start" justifyContent="space-between" gap={1}>
        <Box>
          <Typography component="h2" sx={{ fontSize: '1.125rem', fontWeight: 700 }}>
            {title}
          </Typography>
          <Typography color="textSecondary" sx={{ fontSize: '0.875rem' }}>
            {caption}
          </Typography>
        </Box>
        <Box flexShrink={0}>{action}</Box>
      </Box>
      <Box mt={1.5}>{children}</Box>
    </Paper>
  );
}

function PolicyRowView(props: {
  label: string;
  value: React.ReactNode;
  children?: React.ReactNode;
}) {
  return (
    <Box
      display="flex"
      alignItems="center"
      minHeight="2.5rem"
      borderTop={1}
      borderColor="divider"
      data-testid="policy-row"
    >
      <Box width="40%" pr={1}>
        <CardFeatureLabel>{props.label}</CardFeatureLabel>
      </Box>
      <Box flex={1} display="flex" alignItems="center" flexWrap="wrap" gap={1}>
        <CardLabel>{props.value}</CardLabel>
        {props.children}
      </Box>
    </Box>
  );
}

export interface GroupPolicyPanesProps {
  group: Group;
  instanceMode: Exclude<InstanceMode, 'single'>;
  onEditDefault: () => void;
  onEditOverrides: () => void;
  onPauseUpdates: () => void;
  onResumeUpdates: () => void;
}

// The group default, which the control node replicates to every edge, next to
// what applies on this node. Only a control node edits the default.
export default function GroupPolicyPanes(props: GroupPolicyPanesProps) {
  const { group, instanceMode, onEditDefault, onEditOverrides, onPauseUpdates, onResumeUpdates } =
    props;
  const { t } = useTranslation();
  const rows = policyRows(t);
  const effective = effectivePolicy(group);
  const isControl = instanceMode === 'control';

  const defaultTitle = t('groups|default_policy');
  const localTitle = isControl ? t('groups|this_node_control') : t('groups|this_node_edge');

  return (
    <Grid container spacing={2}>
      <Grid size={{ xs: 12, md: 6 }}>
        <PolicyPane
          title={defaultTitle}
          caption={
            isControl ? t('groups|default_policy_caption') : t('groups|default_policy_caption_edge')
          }
          testId="default-policy-pane"
          action={
            isControl ? (
              <Button
                size="small"
                startIcon={<EditIcon fontSize="small" />}
                onClick={onEditDefault}
                aria-label={t('groups|edit_default_policy')}
              >
                {t('frequent|edit')}
              </Button>
            ) : (
              <Tooltip title={t('groups|default_policy_read_only')}>
                <LockOutlinedIcon
                  color="disabled"
                  titleAccess={t('groups|default_policy_read_only')}
                />
              </Tooltip>
            )
          }
        >
          {rows.map(row => (
            <PolicyRowView key={row.label} label={row.label} value={row.value(group, false)} />
          ))}
        </PolicyPane>
      </Grid>
      <Grid size={{ xs: 12, md: 6 }}>
        <PolicyPane
          title={localTitle}
          caption={t('groups|this_node_caption')}
          testId="local-policy-pane"
          action={
            <Button
              size="small"
              startIcon={<EditIcon fontSize="small" />}
              onClick={onEditOverrides}
              aria-label={t('groups|edit_overrides')}
            >
              {t('frequent|edit')}
            </Button>
          }
        >
          {rows.map(row => (
            <PolicyRowView
              key={row.label}
              label={row.label}
              value={row.value(effective, isPausedHere(group))}
            >
              {isOverridden(group, ...row.fields) && <OverrideLabel />}
              {row.fields.includes('policy_updates_enabled') &&
                (isPausedHere(group) ? (
                  <Button
                    size="small"
                    onClick={onResumeUpdates}
                    aria-label={t('groups|resume_updates_label')}
                  >
                    {t('groups|resume_updates')}
                  </Button>
                ) : (
                  <Button
                    size="small"
                    onClick={onPauseUpdates}
                    aria-label={t('groups|pause_updates_label')}
                  >
                    {t('groups|pause_updates')}
                  </Button>
                ))}
            </PolicyRowView>
          ))}
        </PolicyPane>
      </Grid>
    </Grid>
  );
}
