import PropTypes from 'prop-types';
import { applicationsStore } from "../../stores/Stores"
import React from "react"
import Switch from "rc-switch"
import _ from "underscore"
import ChannelLabel from "../Common/ChannelLabel.react"
import InstancesContainer from "../Instances/Container.react"
import VersionBreakdown from "../Common/VersionBreakdown.react"
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import {CardFeatureLabel, CardHeader} from '../Common/Card';
import Typography from '@material-ui/core/Typography';

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
      groupChannel = _.isEmpty(group.channel) ? "No channel provided" : <ChannelLabel channel={group.channel} />
      styleGroupChannel = _.isEmpty(group.channel) ? "italicText" : ""
    }

		return (
      <Card>
        <CardHeader
          cardMainLinkLabel={group.name}
          cardId={groupId}
          cardDescription={description}
        />
        <CardContent>
          <Grid container spacing={2}>
            <Grid item>
              <CardFeatureLabel>Instances:</CardFeatureLabel>
              <Typography className="activeLink" component="span">{instancesNum}</Typography>
            </Grid>
            <Grid item>
              <CardFeatureLabel>Channel:</CardFeatureLabel>
              <span className={styleGroupChannel}>{groupChannel}</span>
            </Grid>
            <Grid item xs={12}>
              <CardFeatureLabel>Rollout Policy:</CardFeatureLabel>
              Max {policyMaxUpdatesPerDay} updates per {policyPeriodInterval} &nbsp;|&nbsp; Updates timeout { policyUpdatesTimeout }
            </Grid>
            <Grid item xs={4}>
              <CardFeatureLabel>Updates Enabled:</CardFeatureLabel>
              <Switch checked={policyUpdates} disabled={true} checkedChildren={"✔"} unCheckedChildren={"✘"} />
            </Grid>
            <Grid item xs={4}>
              <CardFeatureLabel>Only Office Hours:</CardFeatureLabel>
              <Switch checked={officeHours} disabled={true} checkedChildren={"✔"} unCheckedChildren={"✘"} />
            </Grid>
            <Grid item xs={4}>
              <CardFeatureLabel>Safe Mode:</CardFeatureLabel>
              <Switch checked={safeMode} disabled={true} checkedChildren={"✔"} unCheckedChildren={"✘"} />
            </Grid>
            <Grid item xs={12} className="groups--resume">
              <VersionBreakdown version_breakdown={version_breakdown} channel={channel} />
            </Grid>
            <Grid item xs={12} className="groups--resume">
              {/* Instances */}
              <InstancesContainer
                appID={this.props.appID}
                groupID={this.props.groupID}
                version_breakdown={version_breakdown}
                channel={channel} />
            </Grid>
          </Grid>
        </CardContent>
      </Card>
		)
  }

}

ItemExtended.propTypes = {
  appID: PropTypes.string.isRequired,
  groupID: PropTypes.string.isRequired
}

export default ItemExtended
