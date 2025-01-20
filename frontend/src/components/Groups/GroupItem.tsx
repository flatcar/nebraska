import CheckIcon from '@mui/icons-material/Check';
import CloseIcon from '@mui/icons-material/Close';
import ScheduleIcon from '@mui/icons-material/Schedule';
import { Box, Divider, Typography } from '@mui/material';
import Grid from '@mui/material/Grid';
import makeStyles from '@mui/styles/makeStyles';
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

const useStyles = makeStyles(theme => ({
  root: {
    paddingBottom: 0,
  },
  itemSection: {
    padding: '1em',
  },
  success: {
    color: theme.palette.success.main,
  },
  last24hours: {
    color: 'rgba(0,0,0,0.6)',
    fontSize: '0.875rem',
    lineHeight: 2.0,
  },
}));

// From this number, we stop rate limiting in the backend
const MAX_UPDATES_PER_TIME_PERIOD = 900000;

export function formatUpdateLimits(t: TFunction, group: Group) {
  if (group.policy_max_updates_per_period >= MAX_UPDATES_PER_TIME_PERIOD) {
    return t('groups|Unlimited number of parallel updates');
  }
  return t('groups|Max {{policy_max_updates_per_period, number}} / {{policy_period_interval}}', {
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
    const confirmationText = t('groups|Are you sure you want to delete this group?');
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
  const classes = useStyles();
  const { t } = useTranslation();

  const description = group.description || t('groups|No description provided');
  const channel = group.channel || null;

  const groupChannel = _.isEmpty(group.channel) ? (
    <CardLabel>{t('groups|No channel provided')}</CardLabel>
  ) : (
    <ChannelItem channel={group.channel} />
  );

  return (
    <ListItem disableGutters className={classes.root}>
      <Grid container>
        <Grid item xs={12}>
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
                  label: t('frequent|Edit'),
                  action: () => handleUpdateGroup(group.application_id, group.id),
                },
                {
                  label: t('frequent|Delete'),
                  action: () => deleteGroup(group.application_id, group.id),
                },
              ]}
            />
          </CardHeader>
        </Grid>
        <Grid item xs={12} container justifyContent="space-between">
          <Grid item xs={4} container direction="column" className={classes.itemSection}>
            <Grid item>
              <CardFeatureLabel>{t('groups|Instances')}</CardFeatureLabel>
              <Box>
                <CardLabel labelStyle={{ fontSize: '1.5rem' }}>
                  {totalInstances !== null ? (
                    totalInstances > 0 ? (
                      totalInstances
                    ) : (
                      t('frequent|None')
                    )
                  ) : (
                    <Empty>{t('frequent|Loading...')}</Empty>
                  )}
                </CardLabel>
                <Box display="flex" mr={2}>
                  <ScheduleIcon color="disabled" />
                  <Box pl={1} color="text.disabled">
                    <Typography className={classes.last24hours}>
                      {t('groups|last 24 hours')}
                    </Typography>
                  </Box>
                </Box>
              </Box>
            </Grid>
          </Grid>
          <Box width="1%">
            <Divider orientation="vertical" variant="fullWidth" />
          </Box>
          <Grid item xs={7} container direction="column" className={classes.itemSection}>
            <Grid item>
              <CardFeatureLabel>{t('groups|Channel')}</CardFeatureLabel> {groupChannel}
            </Grid>
            <Grid item>
              <CardFeatureLabel>{t('groups|Updates')}</CardFeatureLabel>
              <Box p={1} mb={1}>
                <CardLabel>
                  <Box display="flex">
                    {group.policy_updates_enabled ? (
                      <>
                        <Box>{t('frequent|Enabled')}</Box>
                        <Box>
                          <CheckIcon className={classes.success} fontSize={'small'} />
                        </Box>
                      </>
                    ) : (
                      <>
                        <Box>{t('frequent|Disabled')}</Box>
                        <Box>
                          <CloseIcon color="error" />
                        </Box>
                      </>
                    )}
                  </Box>
                </CardLabel>
              </Box>
            </Grid>
            <Grid item>
              <CardFeatureLabel>{t('groups|Rollout Policy')}</CardFeatureLabel>
              <Box p={1} mb={1}>
                <CardLabel>{formatUpdateLimits(t, group)}</CardLabel>
              </Box>
            </Grid>
            <Grid item container>
              <Grid item xs={12}>
                <CardFeatureLabel>{t('groups|Version breakdown')}</CardFeatureLabel>
              </Grid>
              <Grid item xs={12}>
                {versionBreakdown === null ? (
                  <Empty>{t('frequent|Loading...')}</Empty>
                ) : versionBreakdown?.length > 0 ? (
                  <VersionProgressBar version_breakdown={versionBreakdown} channel={channel} />
                ) : (
                  <Empty>{t('groups|No instances available.')}</Empty>
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
