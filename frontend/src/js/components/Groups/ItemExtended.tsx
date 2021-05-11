import { Divider } from '@material-ui/core';
import Box from '@material-ui/core/Box';
import Grid from '@material-ui/core/Grid';
import Paper from '@material-ui/core/Paper';
import { makeStyles, useTheme } from '@material-ui/core/styles';
import CheckIcon from '@material-ui/icons/Check';
import CloseIcon from '@material-ui/icons/Close';
import React from 'react';
import { useHistory, useLocation } from 'react-router-dom';
import _ from 'underscore';
import API from '../../api/API';
import { Group } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { defaultTimeInterval, timeIntervalsDefault } from '../../utils/helpers';
import ChannelItem from '../Channels/Item';
import { CardFeatureLabel, CardHeader, CardLabel } from '../Common/Card';
import MoreMenu from '../Common/MoreMenu';
import TimeIntervalLinks from '../Common/TimeIntervalLinks';
import InstanceStatusArea from '../Instances/Charts';
import { StatusCountTimeline, VersionCountTimeline } from './Charts';

const useStyles = makeStyles(theme => ({
  link: {
    fontSize: '1rem',
  },
  instancesChartPaper: {
    height: '100%',
  },
  success: {
    color: theme.palette.success.main,
  },
}));

function ItemExtended(props: {
  appID: string;
  groupID: string;
  handleUpdateGroup: (groupID: string, appID: string) => void;
}) {
  const [application, setApplication] = React.useState(null);
  const [loadingUpdateProgressChart, setLoadingUpdateProgressChart] = React.useState(false);
  const [group, setGroup] = React.useState<Group | null>(null);
  const [instancesStats, setInstancesStats] = React.useState<{
    [key: string]: any;
    total: number;
  } | null>(null);
  const [updateProgressChartDuration, setUpdateProgressChartDuration] = React.useState(
    defaultTimeInterval
  );
  const [versionChartSelectedDuration, setVersionChartSelectedDuration] = React.useState(
    defaultTimeInterval
  );
  const [statusChartDuration, setStatusChartDuration] = React.useState(defaultTimeInterval);
  const location = useLocation();
  const history = useHistory();
  const classes = useStyles();
  const theme = useTheme();
  function onChange() {
    const app = applicationsStore.getCachedApplication(props.appID);

    if (!app) {
      applicationsStore.getApplication(props.appID);
      return;
    }

    if (app !== application) {
      setApplication(app);
    }

    const groupFound = app ? _.findWhere(app.groups, { id: props.groupID }) : null;
    if (groupFound !== group) {
      setGroup(groupFound);
    }
  }
  function updateGroup() {
    props.handleUpdateGroup(props.groupID, props.appID);
  }

  function setDurationToURL(key: string, duration: string) {
    const searchParams = new URLSearchParams(location.search);
    searchParams.set(key, duration);
    history.push({
      pathname: location.pathname,
      search: searchParams.toString(),
    });
  }

  React.useEffect(() => {
    applicationsStore.addChangeListener(onChange);
    onChange();

    return function cleanup() {
      applicationsStore.removeChangeListener(onChange);
    };
  }, []);

  function setDurationStateForCharts(
    key: string,
    setState: React.Dispatch<
      React.SetStateAction<{
        displayValue: string;
        queryValue: string;
        disabled: boolean;
      }>
    >
  ) {
    const searchParams = new URLSearchParams(location.search);
    const period = searchParams.get(key);
    const selectedInterval = timeIntervalsDefault.find(
      intervals => intervals.queryValue === period
    );
    setState(selectedInterval || defaultTimeInterval);
  }

  React.useEffect(() => {
    setDurationStateForCharts('version_timeline_period', setVersionChartSelectedDuration);
    setDurationStateForCharts('status_timeline_period', setStatusChartDuration);
    setDurationStateForCharts('stats_period', setUpdateProgressChartDuration);
  }, [location]);

  React.useEffect(() => {
    if (group) {
      setLoadingUpdateProgressChart(true);
      API.getGroupInstancesStats(
        group.application_id,
        group.id,
        updateProgressChartDuration.queryValue
      )
        .then(stats => {
          setInstancesStats(stats);
          setLoadingUpdateProgressChart(false);
        })
        .catch(err => {
          console.error(
            'Error getting instances stats in Groups/ItemExtended. Group:',
            group,
            '\nError:',
            err
          );
          setInstancesStats(null);
          setLoadingUpdateProgressChart(false);
        });
    }
  }, [group, updateProgressChartDuration]);

  return (
    <Paper>
      <Grid container alignItems="stretch" justify="space-between">
        <Grid item xs={12}>
          <CardHeader
            cardMainLinkLabel={group ? group.name : '…'}
            cardId={group ? group.id : '…'}
            cardTrack={group ? group.track : ''}
            cardDescription={group ? group.description : ''}
          >
            <MoreMenu
              options={[
                {
                  label: 'Edit',
                  action: updateGroup,
                },
              ]}
            />
          </CardHeader>
        </Grid>
        <Grid item xs={4}>
          <Grid container>
            {group && (
              <Grid item xs={12}>
                <Box p={2}>
                  <Grid container direction="column" justify="space-between">
                    <Grid item>
                      <CardFeatureLabel>Channel</CardFeatureLabel>
                      {_.isEmpty(group.channel) ? (
                        <Box my={1}>
                          <CardLabel>No channel assigned</CardLabel>
                        </Box>
                      ) : (
                        <ChannelItem channel={group.channel} />
                      )}
                    </Grid>
                    <Grid item>
                      <CardFeatureLabel>Updates</CardFeatureLabel>
                      <Box my={1}>
                        <CardLabel>
                          <Box display="flex">
                            {group.policy_updates_enabled ? (
                              <>
                                <Box>Enabled</Box>
                                <Box pl={1}>
                                  <CheckIcon className={classes.success} fontSize="small" />
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
                      <CardFeatureLabel>Only Office Hours</CardFeatureLabel>
                      <Box my={1}>
                        <CardLabel>{group.policy_office_hours ? 'Yes' : 'No'}</CardLabel>
                      </Box>
                    </Grid>
                    <Grid item>
                      <CardFeatureLabel>Safe Mode</CardFeatureLabel>
                      <Box my={1}>
                        <CardLabel>{group.policy_safe_mode ? 'Yes' : 'No'}</CardLabel>
                      </Box>
                    </Grid>
                    <Grid item>
                      <CardFeatureLabel>Updates Policy</CardFeatureLabel>
                      <Box my={1}>
                        <CardLabel>
                          {`Max ${group.policy_max_updates_per_period || 0} updates per ${
                            group.policy_period_interval || 0
                          }`}
                        </CardLabel>
                      </Box>
                    </Grid>
                    <Grid item>
                      <CardFeatureLabel>Updates Timeout</CardFeatureLabel>
                      <Box my={1}>
                        <CardLabel>{group.policy_update_timeout}</CardLabel>
                      </Box>
                    </Grid>
                  </Grid>
                </Box>
              </Grid>
            )}
          </Grid>
        </Grid>
        <Box>
          <Divider orientation="vertical" />
        </Box>
        <Grid item xs={7}>
          <Box mt={1} ml={-3}>
            {group && (
              <>
                <Grid container alignItems="center" justify="space-between" spacing={2}>
                  <Grid item>
                    <Box color={theme.palette.greyShadeColor} fontSize={18} fontWeight={700}>
                      Update Progress
                    </Box>
                  </Grid>
                  <Grid item>
                    <Box m={2}>
                      <TimeIntervalLinks
                        intervalChangeHandler={duration =>
                          setDurationToURL('stats_period', duration.queryValue)
                        }
                        selectedInterval={updateProgressChartDuration.queryValue}
                        appID={props.appID}
                        groupID={props.groupID}
                      />
                    </Box>
                  </Grid>
                </Grid>
                <Box padding="1em">
                  <InstanceStatusArea
                    instanceStats={instancesStats}
                    period={updateProgressChartDuration.displayValue}
                    href={{
                      pathname: `/apps/${props.appID}/groups/${props.groupID}/instances`,
                      search: `period=${updateProgressChartDuration.queryValue}`,
                    }}
                  />
                </Box>
              </>
            )}
          </Box>
        </Grid>
        <Grid item xs={12}>
          <Divider variant="fullWidth" />
        </Grid>
        {instancesStats && instancesStats.total > 0 && (
          <Grid item xs={12} container>
            <Grid item md xs={12} container direction="column">
              <Grid container alignItems="center" justify="space-between">
                <Grid item>
                  <Box pl={4} pt={4}>
                    <Box fontSize={18} fontWeight={700} color={theme.palette.greyShadeColor}>
                      Version Breakdown
                    </Box>
                  </Box>
                </Grid>
                <Grid item>
                  <Box pt={4} pr={2}>
                    <TimeIntervalLinks
                      intervalChangeHandler={duration =>
                        setDurationToURL('version_timeline_period', duration.queryValue)
                      }
                      selectedInterval={versionChartSelectedDuration.queryValue}
                      appID={props.appID}
                      groupID={props.groupID}
                    />
                  </Box>
                </Grid>
              </Grid>
              <Box padding="2em">
                <VersionCountTimeline group={group} duration={versionChartSelectedDuration} />
              </Box>
            </Grid>
            <Box width="1%">
              <Divider orientation="vertical" />
            </Box>
            <Grid item md xs={12} container direction="column">
              <Grid container alignItems="center" justify="space-between">
                <Grid item>
                  <Box
                    pl={2}
                    color={theme.palette.greyShadeColor}
                    fontSize={18}
                    fontWeight={700}
                    pt={4}
                  >
                    Status Breakdown
                  </Box>
                </Grid>
                <Grid item>
                  <Box pt={4} pr={2}>
                    <TimeIntervalLinks
                      intervalChangeHandler={duration =>
                        setDurationToURL('status_timeline_period', duration.queryValue)
                      }
                      selectedInterval={statusChartDuration.queryValue}
                      appID={props.appID}
                      groupID={props.groupID}
                    />
                  </Box>
                </Grid>
              </Grid>
              <Box padding="2em">
                <StatusCountTimeline group={group} duration={statusChartDuration} />
              </Box>
            </Grid>
          </Grid>
        )}
      </Grid>
    </Paper>
  );
}

export default ItemExtended;
