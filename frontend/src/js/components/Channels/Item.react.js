import PropTypes from 'prop-types';
import { applicationsStore } from "../../stores/Stores"
import React from "react"
import Grid from '@material-ui/core/Grid';
import ChannelLabel from "../Common/ChannelLabel.react"

class Item extends React.Component {

  constructor(props) {
    super(props)
    this.deleteChannel = this.deleteChannel.bind(this)
    this.updateChannel = this.updateChannel.bind(this)
  }

  deleteChannel() {
    let confirmationText = "Are you sure you want to delete this channel?"
    if (confirm(confirmationText)) {
      applicationsStore.deleteChannel(this.props.channel.application_id, this.props.channel.id)
    }
  }

  updateChannel() {
    this.props.handleUpdateChannel(this.props.channel.id)
  }

  render() {
    let popoverContent = {
      type: "channel",
      appID: this.props.channel.application_id,
      channelID: this.props.channel.id
    }

    const name = this.props.channel ? this.props.channel.name : "",
          version = (this.props.channel && this.props.channel.package) ? this.props.channel.package.version : "-"

    return (
      <Grid container>
        <Grid item xs={8}>
          <ChannelLabel channel={this.props.channel} channelLabelStyle="fixedWidth" />
        </Grid>
        <Grid item xs={4} className="alignRight">
          <div className="channelsList-buttons">
            <button className="cr-button displayInline fa fa-edit" onClick={this.updateChannel}></button>
            <button className="cr-button displayInline fa fa-trash-o" onClick={this.deleteChannel}></button>
          </div>
        </Grid>
      </Grid>
    )
  }

}

Item.propTypes = {
  channel: PropTypes.object.isRequired,
  packages: PropTypes.array.isRequired,
  handleUpdateChannel: PropTypes.func.isRequired
}

export default Item
