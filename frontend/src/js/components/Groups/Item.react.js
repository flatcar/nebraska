import Grid from '@material-ui/core/Grid';
import Link from '@material-ui/core/Link';
import { makeStyles } from '@material-ui/core/styles';
import PropTypes from 'prop-types';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import _ from 'underscore';
import { applicationsStore } from '../../stores/Stores';
import ChannelItem from '../Channels/Item.react';
import { CardFeatureLabel, CardHeader } from '../Common/Card';
import ListItem from '../Common/ListItem';
import MoreMenu from '../Common/MoreMenu';
import VersionBreakdown from '../Common/VersionBreakdown.react';

const useStyles = makeStyles(theme => ({
  itemSection: {
    padding: '1em'
  },
}));

function Item(props) {
  const classes = useStyles();

  let version_breakdown = (props.group && props.group.version_breakdown) ? props.group.version_breakdown : [];
  let instances_total = props.group.instances_stats ? props.group.instances_stats.total : 0;
  let description = props.group.description || 'No description provided';
  let channel = props.group.channel || {};

  let groupChannel = _.isEmpty(props.group.channel) ? "No channel provided"
    : <ChannelItem channel={props.group.channel} ContainerComponent="span" />
  let styleGroupChannel = _.isEmpty(props.group.channel) ? "italicText" : ""
  let groupPath = `/apps/${props.group.application_id}/groups/${props.group.id}`

  function deleteGroup() {
    let confirmationText = "Are you sure you want to delete this group?"
    if (confirm(confirmationText)) {
      applicationsStore.deleteGroup(props.group.application_id, props.group.id)
    }
  }

  function updateGroup() {
    props.handleUpdateGroup(props.group.application_id, props.group.id)
  }

  return (
    <ListItem disableGutters>
      <Grid
        container
      >
        <Grid item xs={12}>
          <CardHeader
            cardMainLinkLabel={props.group.name}
            cardMainLinkPath={groupPath}
            cardId={props.group.id}
            cardDescription={description}
          >
            <MoreMenu options={[
              {
              'label': 'Edit',
              'action': updateGroup,
              },
              {
                'label': 'Delete',
                'action': deleteGroup,
              }
            ]} />
          </CardHeader>
        </Grid>
        <Grid
          item
          xs={12}
          container
          justify="space-between"
          className={classes.itemSection}
        >
          <Grid item xs={6} container direction="column">
            <Grid item>
              <CardFeatureLabel>Instances:</CardFeatureLabel>
              <Link
                to={groupPath}
                component={RouterLink}
              >
                {instances_total}
              </Link>
            </Grid>
            <Grid item>
              <CardFeatureLabel>Channel:</CardFeatureLabel>
              {groupChannel}
            </Grid>
          </Grid>
          <Grid item xs={6} container direction="column">
            <Grid item>
              <CardFeatureLabel>Updates:</CardFeatureLabel>&nbsp;
              {props.group.policy_updates_enabled ? 'Enabled' : 'Disabled'}
            </Grid>
            <Grid item>
              <CardFeatureLabel>Rollout Policy:</CardFeatureLabel>&nbsp;
              Max {props.group.policy_max_updates_per_period} updates per {props.group.policy_period_interval}
            </Grid>
          </Grid>
          <Grid item xs={12}>
            <VersionBreakdown version_breakdown={version_breakdown} channel={channel} />
          </Grid>
        </Grid>
      </Grid>
    </ListItem>
  );
}

Item.propTypes = {
    group: PropTypes.object.isRequired,
    appName: PropTypes.string.isRequired,
    channels: PropTypes.array.isRequired,
    handleUpdateGroup: PropTypes.func.isRequired
}

export default Item
