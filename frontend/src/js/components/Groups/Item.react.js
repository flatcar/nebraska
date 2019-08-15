import PropTypes from 'prop-types';
import { applicationsStore } from "../../stores/Stores"
import React from "react";
import { Link } from "react-router-dom";
import Switch from "rc-switch"
import _ from "underscore"
import ChannelLabel from "../Common/ChannelLabel.react"
import VersionBreakdown from "../Common/VersionBreakdown.react"
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import {CardFeatureLabel, CardHeader} from '../Common/Card';

class Item extends React.Component {

  constructor(props) {
    super(props)
    this.deleteGroup = this.deleteGroup.bind(this)
    this.updateGroup = this.updateGroup.bind(this)
  }

  deleteGroup() {
    let confirmationText = "Are you sure you want to delete this group?"
    if (confirm(confirmationText)) {
      applicationsStore.deleteGroup(this.props.group.application_id, this.props.group.id)
    }
  }

  updateGroup() {
    this.props.handleUpdateGroup(this.props.group.application_id, this.props.group.id)
  }

  render() {
    let version_breakdown = (this.props.group && this.props.group.version_breakdown) ? this.props.group.version_breakdown : [],
        instances_total = this.props.group.instances_stats ? this.props.group.instances_stats.total : 0,
        description = this.props.group.description ? this.props.group.description : "No description provided",
        channel = this.props.group.channel ? this.props.group.channel : {}

    let groupChannel = _.isEmpty(this.props.group.channel) ? "No channel provided" : <ChannelLabel channel={this.props.group.channel} />
    let styleGroupChannel = _.isEmpty(this.props.group.channel) ? "italicText" : ""
    let groupPath = `/apps/${this.props.group.application_id}/groups/${this.props.group.id}`

    return (
      <Card>
        <CardHeader
          cardMainLinkLabel={this.props.group.name}
          cardMainLinkPath={groupPath}
          cardId={this.props.group.id}
          cardDescription={description}
        >
          <div className="apps--buttons">
            <button className="cr-button displayInline fa fa-edit" onClick={this.updateGroup}></button>
            <button className="cr-button displayInline fa fa-trash-o" onClick={this.deleteGroup}></button>
          </div>
        </CardHeader>
        <CardContent>
          <Grid container spacing={2}>
            <Grid item>
              <CardFeatureLabel>Instances:</CardFeatureLabel>
              <Link to={groupPath}><span className="activeLink"> {instances_total}<span className="fa fa-caret-right" /></span></Link>
            </Grid>
            <Grid item>
              <CardFeatureLabel>Channel:</CardFeatureLabel>
              {groupChannel}
            </Grid>
            <Grid item xs={8}>
              <CardFeatureLabel>Rollout Policy:</CardFeatureLabel>
              Max {this.props.group.policy_max_updates_per_period} updates per {this.props.group.policy_period_interval}
            </Grid>
            <Grid item xs={4}>
              <CardFeatureLabel>Updates Enabled:</CardFeatureLabel>
              <Switch checked={this.props.group.policy_updates_enabled} disabled={true} checkedChildren={"✔"} unCheckedChildren={"✘"} />
            </Grid>
            <Grid item xs={12}>
              <VersionBreakdown version_breakdown={version_breakdown} channel={channel} />
            </Grid>
          </Grid>
        </CardContent>
      </Card>
    )
  }

}

Item.propTypes = {
    group: PropTypes.object.isRequired,
    appName: PropTypes.string.isRequired,
    channels: PropTypes.array.isRequired,
    handleUpdateGroup: PropTypes.func.isRequired
}


export default Item
