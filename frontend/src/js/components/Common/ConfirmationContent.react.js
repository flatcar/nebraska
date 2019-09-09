import PropTypes from 'prop-types';
import React from "react"
import ApplicationsStore from "../../stores/ApplicationsStore"

class ConfirmationContent extends React.Component {

  constructor(props) {
    super(props)
    this.processClick = this.processClick.bind(this)
  }

  processClick() {
    if (this.props.data.type == "application") {
      ApplicationsStore.deleteApplication(this.props.data.appID)
    } else if (this.props.data.type == "group") {
      ApplicationsStore.deleteGroup(this.props.data.appID, this.props.data.groupID)
    } else if (this.props.data.type == "channel") {
      ApplicationsStore.deleteChannel(this.props.data.appID, this.props.data.channelID)
    } else if (this.props.data.type == "package") {
      ApplicationsStore.deletePackage(this.props.data.appID, this.props.data.packageID)
    }
  }

  render() {
    return (
      <div className="popover-content" {...this.props}>
        Are you sure ... ?
        <p className="button-group">
          <button type="button" className="confirm-dialog-btn-abord">No</button>
          <button type="button" className="confirm-dialog-btn-confirm" onClick={this.processClick}>Yes</button>
        </p>
      </div>
    )
  }

}

ConfirmationContent.propTypes = {
  channel: PropTypes.object.isRequired,
  data: PropTypes.object.isRequired
}

export default ConfirmationContent
