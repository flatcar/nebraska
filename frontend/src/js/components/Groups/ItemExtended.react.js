import Box from '@material-ui/core/Box';
import Grid from '@material-ui/core/Grid';
import Link from '@material-ui/core/Link';
import Paper from '@material-ui/core/Paper';
import { makeStyles } from '@material-ui/core/styles';
import PropTypes from 'prop-types';
import React from "react";
import { Link as RouterLink } from 'react-router-dom';
import _ from "underscore";
import { applicationsStore } from "../../stores/Stores";
import ChannelItem from '../Channels/Item.react';
import { CardFeatureLabel, CardHeader, CardLabel } from '../Common/Card';
import InstanceStatusArea from '../Instances/Charts';
import GroupTimelineChart from './Charts';
import ListHeader from '../Common/ListHeader';

const useStyles = makeStyles({
  link: {
    fontSize: '1rem'
  },
  instancesChartPaper: {
    height: '100%',
  },
});

function AllInstancesButton(props) {
  const classes = useStyles();
  let {path} = props;

  return (
    <Link
      className={classes.link}
      to={{pathname: path}}
      component={RouterLink}
    >
      See instances
    </Link>
  );
}

function ItemExtended(props) {
  const [application, setApplication] = React.useState(null);
  const [group, setGroup] = React.useState(null);
  const classes = useStyles();

  function onChange() {
    let app = applicationsStore.getCachedApplication(props.appID);

    if (!app) {
      applicationsStore.getApplication(props.appID);
      return;
    }

    if (app !== application) {
      setApplication(app);
    }

    let groupFound = app ? _.findWhere(app.groups, {id: props.groupID}) : null;
    if (groupFound !== group) {
      setGroup(groupFound);
    }
  }

  React.useEffect(() => {
    applicationsStore.addChangeListener(onChange);
    onChange();

    return function cleanup() {
      applicationsStore.removeChangeListener(onChange);
    }
  },
  [application, group]);

  return (
    <Grid
      container
      spacing={2}
      alignItems="stretch"
    >
      <Grid item xs={5}>
        <Paper>
          <Grid container>
            <Grid item xs={12}>
              <CardHeader
                cardMainLinkLabel={group ? group.name : '…'}
                cardId={group ? group.id : '…'}
                cardDescription={group ? group.description : ''}
              />
            </Grid>
            {group &&
              <Grid item xs={12}>
                <Box padding="1em">
                  <Grid
                    container
                    direction="column"
                    justify="space-between"
                    spacing={1}
                  >
                    <Grid item>
                      <CardFeatureLabel>Channel:</CardFeatureLabel>
                      {_.isEmpty(group.channel) ?
                        <CardLabel>No channel provided</CardLabel>
                      :
                        <ChannelItem
                          channel={group.channel}
                          ContainerComponent="span"
                        />
                      }
                    </Grid>
                    <Grid item>
                      <CardFeatureLabel>Updates:</CardFeatureLabel>&nbsp;
                      <CardLabel>{group.policy_updates_enabled ? 'Enabled' : 'Disabled'}</CardLabel>
                    </Grid>
                    <Grid item>
                      <CardFeatureLabel>Only Office Hours:</CardFeatureLabel>&nbsp;
                      <CardLabel>{group.policy_office_hours ? 'Yes' : 'No'}</CardLabel>
                    </Grid>
                    <Grid item>
                      <CardFeatureLabel>Safe Mode:</CardFeatureLabel>&nbsp;
                      <CardLabel>{group.policy_safe_mode ? 'Yes' : 'No'}</CardLabel>
                    </Grid>
                    <Grid item>
                      <CardFeatureLabel>Updates Policy:</CardFeatureLabel>&nbsp;
                      <CardLabel>Max {group.policy_max_updates_per_period || 0} updates per {group.policy_period_interval || 0}</CardLabel>
                    </Grid>
                    <Grid item>
                      <CardFeatureLabel>Updates Timeout:</CardFeatureLabel>&nbsp;
                      <CardLabel>{group.policy_update_timeout}</CardLabel>
                    </Grid>
                  </Grid>
                </Box>
              </Grid>
            }
          </Grid>
        </Paper>
      </Grid>
      <Grid item xs={7}>
        {group &&
          <Paper className={classes.instancesChartPaper}>
            <ListHeader
              title="Update Progress"
            />
            <Box padding="1em">
              <InstanceStatusArea instanceStats={group.instances_stats} />
            </Box>
          </Paper>
        }
      </Grid>
      { (group && group.instances_stats.total > 0) &&
        <Grid item xs={12}>
          <Paper>
            <ListHeader title="Version Breakdown" />
            <Box padding="1em">
              <Grid
                container
                spacing={3}
                alignItems="center"
                direction="column"
              >
                <Grid item xs={12}>
                  <GroupTimelineChart group={group} />
                </Grid>
                <Grid item xs={12}>
                  <AllInstancesButton path={`/apps/${props.appID}/groups/${props.groupID}/instances`} />
                </Grid>
              </Grid>
            </Box>
          </Paper>
        </Grid>
      }
    </Grid>
  );
}

ItemExtended.propTypes = {
  appID: PropTypes.string.isRequired,
  groupID: PropTypes.string.isRequired
}

export default ItemExtended
