import CheckIcon from '@mui/icons-material/Check';
import CloseIcon from '@mui/icons-material/Close';
import ScheduleIcon from '@mui/icons-material/Schedule';
import { Box, Divider, Typography } from '@mui/material';
import Grid from '@mui/material/Grid';
import { TFunction } from 'i18next';
import React from 'react';
import { useTranslation } from 'react-i18next';
import _ from 'underscore';

import API from '../../api/API';
import { Group, VersionBreakdownEntry } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { useGroupVersionBreakdown } from '../../utils/helpers';
import ChannelItem from '../Channels/ChannelItem';
import { CardFeatureLabel, CardHeader, CardLabel } from '../common/Card';
import Empty from '../common/EmptyContent';
import ListItem from '../common/ListItem';
import MoreMenu from '../common/MoreMenu';
import VersionProgressBar from '../common/VersionBreakdownBar';

// From this number, we stop rate limiting in the backend
const MAX_UPDATES_PER_TIME_PERIOD = 900000;

export function formatUpdateLimits(t: TFunction, group: Group) {
  if (group.policy_max_updates_per_period >= MAX_UPDATES_PER_TIME_PERIOD) {
    return t('groups|update_policy_unlimited');
  }
  return t('groups|max_updates_per_period', {
    policy_max_updates_per_period: group.policy_max_updates_per_period,
    policy_period_interval: group.policy_period_interval,
  });
}

export interface GroupItemProps {
  group: Group;
  handleUpdateGroup: (appID: string, groupID: string) => void;
}

function GroupItem({ group, handleUpdateGroup }: GroupItemProps) {
  const { t } = useTranslation();
  const [totalInstances, setTotalInstances] = React.useState<null | number>(null);
  const versionBreakdown = useGroupVersionBreakdown(group);

  function deleteGroup(appID: string, groupID: string) {
    const confirmationText = t('groups|group_delete_confirmation');
    if (window.confirm(confirmationText)) {
      applicationsStore().deleteGroup(appID, groupID);
    }
  }

  React.useEffect(() => {
    API.getInstancesCount(group.application_id, group.id, '1d')
      .then(result => {
        setTotalInstances(result);
      })
      .catch(err => console.error('Error getting total instances in Group/Item', err));
  }, []);

  return (
    <PureGroupItem
      group={group}
      handleUpdateGroup={handleUpdateGroup}
      deleteGroup={deleteGroup}
      versionBreakdown={versionBreakdown}
      totalInstances={totalInstances}
    />
  );
}

export interface PureGroupItemProps {
  group: Group;
  versionBreakdown: VersionBreakdownEntry[] | null;
  totalInstances: number | null;
  handleUpdateGroup: (appID: string, groupID: string) => void;
  deleteGroup: (appID: string, groupID: string) => void;
}

export function PureGroupItem({
  group,
  versionBreakdown,
  totalInstances,
  handleUpdateGroup,
  deleteGroup,
}: PureGroupItemProps) {
  const { t } = useTranslation();

  const description = group.description || t('groups|description_none_provided');
  const channel = group.channel || null;

  const groupChannel = _.isEmpty(group.channel) ? (
    <CardLabel>{t('groups|channel_none_provided')}</CardLabel>
  ) : (
    <ChannelItem channel={group.channel} />
  );

  return (
    <ListItem
      disableGutters
      sx={{
        paddingBottom: 0,
      }}
    >
      <Grid container>
        <Grid size={12}>
          <CardHeader
            cardMainLinkLabel={group.name}
            cardMainLinkPath={`/apps/${group.application_id}/groups/${group.id}`}
            cardId={group.id}
            cardTrack={group.track}
            cardDescription={description}
          >
            <MoreMenu
              options={[
                {
                  label: t('frequent|edit'),
                  action: () => handleUpdateGroup(group.application_id, group.id),
                },
                {
                  label: t('frequent|delete'),
                  action: () => deleteGroup(group.application_id, group.id),
                },
              ]}
            />
          </CardHeader>
        </Grid>
        <Grid container justifyContent="space-between" size={12}>
          <Grid container direction="column" sx={{ padding: '1em' }} size={4}>
            <Grid>
              <CardFeatureLabel>{t('groups|instances')}</CardFeatureLabel>
              <Box>
                <CardLabel labelStyle={{ fontSize: '1.5rem' }}>
                  {totalInstances !== null ? (
                    totalInstances > 0 ? (
                      totalInstances
                    ) : (
                      t('frequent|none')
                    )
                  ) : (
                    <Empty>{t('frequent|loading')}</Empty>
                  )}
                </CardLabel>
                <Box display="flex" mr={2}>
                  <ScheduleIcon color="disabled" />
                  <Box pl={1} color="text.disabled">
                    <Typography
                      sx={{
                        color: 'rgba(0,0,0,0.6)',
                        fontSize: '0.875rem',
                        lineHeight: 2.0,
                      }}
                    >
                      {t('groups|time_last_24_hours')}
                    </Typography>
                  </Box>
                </Box>
              </Box>
            </Grid>
          </Grid>
          <Box width="1%">
            <Divider orientation="vertical" variant="fullWidth" />
          </Box>
          <Grid container direction="column" sx={{ padding: '1em' }} size={7}>
            <Grid>
              <CardFeatureLabel>{t('groups|channel')}</CardFeatureLabel> {groupChannel}
            </Grid>
            <Grid>
              <CardFeatureLabel>{t('groups|updates')}</CardFeatureLabel>
              <Box p={1} mb={1}>
                <CardLabel>
                  <Box display="flex">
                    {group.policy_updates_enabled ? (
                      <>
                        <Box>{t('frequent|enabled')}</Box>
                        <Box>
                          <CheckIcon
                            sx={{ color: theme => theme.palette.success.main }}
                            fontSize={'small'}
                          />
                        </Box>
                      </>
                    ) : (
                      <>
                        <Box>{t('frequent|disabled')}</Box>
                        <Box>
                          <CloseIcon color="error" />
                        </Box>
                      </>
                    )}
                  </Box>
                </CardLabel>
              </Box>
            </Grid>
            <Grid>
              <CardFeatureLabel>{t('groups|rollout_policy')}</CardFeatureLabel>
              <Box p={1} mb={1}>
                <CardLabel>{formatUpdateLimits(t, group)}</CardLabel>
              </Box>
            </Grid>
            <Grid container>
              <Grid size={12}>
                <CardFeatureLabel>{t('groups|version_breakdown_lower')}</CardFeatureLabel>
              </Grid>
              <Grid size={12}>
                {versionBreakdown === null ? (
                  <Empty>{t('frequent|loading')}</Empty>
                ) : versionBreakdown?.length > 0 ? (
                  <VersionProgressBar version_breakdown={versionBreakdown} channel={channel} />
                ) : (
                  <Empty>{t('groups|instances_none_available')}</Empty>
                )}
              </Grid>
            </Grid>
          </Grid>
        </Grid>
      </Grid>
    </ListItem>
  );
}

export default GroupItem;
