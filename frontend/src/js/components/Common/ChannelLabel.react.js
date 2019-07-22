import PropTypes from 'prop-types';
import React from "react"
import { cleanSemverVersion } from "../../constants/helpers"

class ChannelLabel extends React.Component {

  constructor() {
    super()
  }

  render() {
    var channelLabelStyle = this.props.channelLabelStyle ? this.props.channelLabelStyle : ""
    var color = this.props.channel ? this.props.channel.color : "#777777"
    var divColor = {
      backgroundColor: color
    }

    var name = this.props.channel ? this.props.channel.name : ""
    var version = (this.props.channel && this.props.channel.package) ? cleanSemverVersion(this.props.channel.package.version) : "-"

    return (
      <div className={"channelLabel " + channelLabelStyle}>
        <div className="colouredCircle" style={divColor}></div>
        <div className="channelName">{name}</div>
        <span className="channelLabel--span">{version}</span>
      </div>
    )
  }

}

ChannelLabel.propTypes = {
  channel: PropTypes.object.isRequired,
  channelLabelStyle: PropTypes.string
}

export default ChannelLabel
