import { Box, Divider, Typography } from '@material-ui/core';
import Grid from '@material-ui/core/Grid';
import { makeStyles } from '@material-ui/core/styles';
import CheckIcon from '@material-ui/icons/Check';
import CloseIcon from '@material-ui/icons/Close';
import ScheduleIcon from '@material-ui/icons/Schedule';
import React from 'react';
import { useTranslation } from 'react-i18next';
import _ from 'underscore';
import API from '../../api/API';
import { Channel, Group } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { useGroupVersionBreakdown } from '../../utils/helpers';
import ChannelItem from '../Channels/Item';
import { CardFeatureLabel, CardHeader, CardLabel } from '../Common/Card';
import Empty from '../Common/EmptyContent';
import ListItem from '../Common/ListItem';
import MoreMenu from '../Common/MoreMenu';
import VersionProgressBar from '../Common/VersionBreakdownBar';

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
}));

function Item(props: {
  group: Group;
  appName: string;
  channels: Channel[];
  handleUpdateGroup: (appID: string, groupID: string) => void;
}) {
  const classes = useStyles();
  const { t } = useTranslation();
  const [totalInstances, setTotalInstances] = React.useState(-1);

  const version_breakdown = useGroupVersionBreakdown(props.group);
  const description = props.group.description || t('groups|No description provided');
  const channel = props.group.channel || null;

  const groupChannel = _.isEmpty(props.group.channel) ? (
    <CardLabel>{t('groups|No channel provided')}</CardLabel>
  ) : (
    <ChannelItem channel={props.group.channel} />
  );
  const styleGroupChannel = _.isEmpty(props.group.channel) ? 'italicText' : '';
  const groupPath = `/apps/${props.group.application_id}/groups/${props.group.id}`;

  function deleteGroup() {
    const confirmationText = t('groups|Are you sure you want to delete this group?');
    if (window.confirm(confirmationText)) {
      applicationsStore.deleteGroup(props.group.application_id, props.group.id);
    }
  }

  function updateGroup() {
    props.handleUpdateGroup(props.group.application_id, props.group.id);
  }

  React.useEffect(() => {
    API.getInstancesCount(props.group.application_id, props.group.id, '1d')
      .then(result => {
        setTotalInstances(result);
      })
      .catch(err => console.error('Error getting total instances in Group/Item', err));
  }, []);

  return (
    <ListItem disableGutters className={classes.root}>
      <Grid container>
        <Grid item xs={12}>
          <CardHeader
            cardMainLinkLabel={props.group.name}
            cardMainLinkPath={groupPath}
            cardId={props.group.id}
            cardTrack={props.group.track}
            cardDescription={description}
          >
            <MoreMenu
              options={[
                {
                  label: t('frequent|Edit'),
                  action: updateGroup,
                },
                {
                  label: t('frequent|Delete'),
                  action: deleteGroup,
                },
              ]}
            />
          </CardHeader>
        </Grid>
        <Grid item xs={12} container justify="space-between">
          <Grid item xs={4} container direction="column" className={classes.itemSection}>
            <Grid item>
              <CardFeatureLabel>{t('groups|Instances')}</CardFeatureLabel>
              <Box>
                <CardLabel labelStyle={{ fontSize: '1.5rem' }}>
                  {totalInstances > 0 ? totalInstances : t('frequent|None')}
                </CardLabel>
                <Box display="flex" mr={2}>
                  <ScheduleIcon color="disabled" />
                  <Box pl={1} color="text.disabled">
                    <Typography variant="subtitle1">{t('groups|last 24 hours')}</Typography>
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
              <CardFeatureLabel>{t('groups|Channel')}</CardFeatureLabel>
              {groupChannel}
            </Grid>
            <Grid item>
              <CardFeatureLabel>{t('groups|Updates')}</CardFeatureLabel>
              <Box p={1} mb={1}>
                <CardLabel>
                  <Box display="flex">
                    {props.group.policy_updates_enabled ? (
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
                <CardLabel>
                  {t(
                    'groups|Max {{policy_max_updates_per_period, number}} / {{policy_period_interval, number}}',
                    {
                      policy_max_updates_per_period: props.group.policy_max_updates_per_period,
                      policy_period_interval: props.group.policy_period_interval,
                    }
                  )}
                </CardLabel>
              </Box>
            </Grid>
            <Grid item container>
              <Grid item xs={12}>
                <CardFeatureLabel>{t('groups|Version breakdown')}</CardFeatureLabel>
              </Grid>
              <Grid item xs={12}>
                {version_breakdown.length > 0 ? (
                  <VersionProgressBar version_breakdown={version_breakdown} channel={channel} />
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

export default Item;
