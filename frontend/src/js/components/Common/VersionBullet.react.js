import PropTypes from 'prop-types';
import React from "react"

class VersionBullet extends React.Component {

  constructor() {
    super()
  }

  render() {
    var divColor = {
      backgroundColor: this.props.channel.color
    }

    return(
      <div className="versionBullet" style={divColor}></div>
    )
  }

}

VersionBullet.propTypes = {
  channel: PropTypes.object.isRequired
}

export default VersionBullet
