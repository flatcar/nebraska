import { Box, Divider, Typography } from '@material-ui/core';
import Grid from '@material-ui/core/Grid';
import { makeStyles } from '@material-ui/core/styles';
import CheckIcon from '@material-ui/icons/Check';
import CloseIcon from '@material-ui/icons/Close';
import ScheduleIcon from '@material-ui/icons/Schedule';
import React from 'react';
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
  const [totalInstances, setTotalInstances] = React.useState(-1);

  const version_breakdown = useGroupVersionBreakdown(props.group);
  const description = props.group.description || 'No description provided';
  const channel = props.group.channel || {};

  const groupChannel = _.isEmpty(props.group.channel) ? (
    <CardLabel>No channel provided</CardLabel>
  ) : (
    <ChannelItem channel={props.group.channel} />
  );
  const styleGroupChannel = _.isEmpty(props.group.channel) ? 'italicText' : '';
  const groupPath = `/apps/${props.group.application_id}/groups/${props.group.id}`;

  function deleteGroup() {
    const confirmationText = 'Are you sure you want to delete this group?';
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
                  label: 'Edit',
                  action: updateGroup,
                },
                {
                  label: 'Delete',
                  action: deleteGroup,
                },
              ]}
            />
          </CardHeader>
        </Grid>
        <Grid item xs={12} container justify="space-between">
          <Grid item xs={4} container direction="column" className={classes.itemSection}>
            <Grid item>
              <CardFeatureLabel>Instances</CardFeatureLabel>
              <Box>
                <CardLabel labelStyle={{ fontSize: '1.5rem' }}>
                  {totalInstances > 0 ? totalInstances : 'None'}
                </CardLabel>
                <Box display="flex" mr={2}>
                  <ScheduleIcon color="disabled" />
                  <Box pl={1} color="text.disabled">
                    <Typography variant="subtitle1">last 24 hours</Typography>
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
              <CardFeatureLabel>Channel</CardFeatureLabel>
              {groupChannel}
            </Grid>
            <Grid item>
              <CardFeatureLabel>Updates</CardFeatureLabel>
              <Box p={1} mb={1}>
                <CardLabel>
                  <Box display="flex">
                    {props.group.policy_updates_enabled ? (
                      <>
                        <Box>Enabled</Box>
                        <Box>
                          <CheckIcon className={classes.success} fontSize={'small'} />
                        </Box>
                      </>
                    ) : (
                      <>
                        <Box>Disabled</Box>
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
              <CardFeatureLabel>Rollout Policy</CardFeatureLabel>
              <Box p={1} mb={1}>
                <CardLabel>
                  {`Max ${props.group.policy_max_updates_per_period} updates per ${props.group.policy_period_interval}`}
                </CardLabel>
              </Box>
            </Grid>
            <Grid item container>
              <Grid item xs={12}>
                <CardFeatureLabel>Version breakdown</CardFeatureLabel>
              </Grid>
              <Grid item xs={12}>
                {version_breakdown.length > 0 ? (
                  <VersionProgressBar version_breakdown={version_breakdown} channel={channel} />
                ) : (
                  <Empty>No instances available.</Empty>
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
