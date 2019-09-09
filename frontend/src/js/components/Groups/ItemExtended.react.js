import PropTypes from 'prop-types';
import { applicationsStore } from "../../stores/Stores"
import React from "react"
import _ from "underscore"
import ChannelItem from '../Channels/Item.react';
import InstancesContainer from "../Instances/Container.react"
import VersionBreakdown from "../Common/VersionBreakdown.react"
import Grid from '@material-ui/core/Grid';
import {CardLabel, CardFeatureLabel, CardHeader} from '../Common/Card';
import Typography from '@material-ui/core/Typography';
import Paper from '@material-ui/core/Paper';
import Box from '@material-ui/core/Box';

class ItemExtended extends React.Component {

  constructor() {
    super()
    this.onChange = this.onChange.bind(this)

    this.state = {applications: applicationsStore.getCachedApplications()}
  }

  componentDidMount() {
    applicationsStore.addChangeListener(this.onChange)
  }

  componentWillUnmount() {
    applicationsStore.removeChangeListener(this.onChange)
  }

  onChange() {
    this.setState({
      applications: applicationsStore.getCachedApplications()
    })
  }

  render() {
    let application = _.findWhere(this.state.applications, {id: this.props.appID})
    let group = application ? _.findWhere(application.groups, {id: this.props.groupID}) : null

    let name = "",
        groupId = "",
        description = "",
        instancesNum = 0,
        policyMaxUpdatesPerDay = 0,
        policyPeriodInterval = 0,
        channel = {},
        version_breakdown = [],
        policyUpdates,
        policyUpdatesTimeout,
        safeMode,
        officeHours,
        groupChannel,
        styleGroupChannel

    if (group) {
      name = group.name
      groupId = group.id
      description = group.description ? group.description : ""
      channel = group.channel ? group.channel : {}
      instancesNum = group.instances_stats ? group.instances_stats.total : 0
      policyMaxUpdatesPerDay = group.policy_max_updates_per_period ? group.policy_max_updates_per_period : 0
      policyPeriodInterval = group.policy_period_interval ? group.policy_period_interval : 0
      policyUpdates = group.policy_updates_enabled ? group.policy_updates_enabled : null
      policyUpdatesTimeout = group.policy_update_timeout ? group.policy_update_timeout : null
      safeMode = group.policy_safe_mode ? group.policy_safe_mode : null
      officeHours = group.policy_office_hours ? group.policy_office_hours : null
      version_breakdown = group.version_breakdown ? group.version_breakdown : []
      groupChannel = _.isEmpty(group.channel) ? "No channel provided"
        : <ChannelItem channel={group.channel} ContainerComponent="span" />
      styleGroupChannel = _.isEmpty(group.channel) ? "italicText" : ""
    }

		return (
      <Paper>
        <Grid
          container
        >
          <Grid item xs={12}>
            <CardHeader
              cardMainLinkLabel={name}
              cardId={groupId}
              cardDescription={description}
            />
          </Grid>
        </Grid>
        <Box padding="1em">
          <Grid item xs={12} container justify="space-between" spacing={1}>
            <Grid item xs={6} container spacing={1} direction="column">
              <Grid item>
                <CardFeatureLabel>Instances:</CardFeatureLabel>&nbsp;
                <CardLabel>{instancesNum ? instancesNum : 'None'}</CardLabel>
              </Grid>
              <Grid item>
                <CardFeatureLabel>Channel:</CardFeatureLabel>
                {groupChannel}
              </Grid>
            </Grid>
            <Grid item xs={6} container spacing={1} direction="column">
              <Grid item>
                <CardFeatureLabel>Updates:</CardFeatureLabel>&nbsp;
                <CardLabel>{policyUpdates ? 'Enabled' : 'Disabled'}</CardLabel>
              </Grid>
              <Grid item>
                <CardFeatureLabel>Only Office Hours:</CardFeatureLabel>&nbsp;
                <CardLabel>{officeHours ? 'Yes' : 'No'}</CardLabel>
              </Grid>
              <Grid item>
                <CardFeatureLabel>Safe Mode:</CardFeatureLabel>&nbsp;
                <CardLabel>{safeMode ? 'Yes' : 'No'}</CardLabel>
              </Grid>
              <Grid item>
                <CardFeatureLabel>Updates Policy:</CardFeatureLabel>&nbsp;
                <CardLabel>Max {policyMaxUpdatesPerDay} updates per {policyPeriodInterval}</CardLabel>
              </Grid>
              <Grid item>
                <CardFeatureLabel>Updates Timeout:</CardFeatureLabel>&nbsp;
                <CardLabel>Updates timeout { policyUpdatesTimeout }</CardLabel>
              </Grid>
            </Grid>
            <Grid item xs={12}>
              <VersionBreakdown version_breakdown={version_breakdown} channel={channel} />
            </Grid>
            <Grid item xs={12}>
              {/* Instances */}
              <InstancesContainer
                appID={this.props.appID}
                groupID={this.props.groupID}
                version_breakdown={version_breakdown}
                channel={channel} />
            </Grid>
          </Grid>
        </Box>
      </Paper>
		)
  }

}

ItemExtended.propTypes = {
  appID: PropTypes.string.isRequired,
  groupID: PropTypes.string.isRequired
}

export default ItemExtended
