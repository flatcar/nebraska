import PropTypes from 'prop-types';
import React from "react"
import Item from "./ApplicationItemGroupItem.react"

class ApplicationItemGroupsList extends React.Component {

  constructor() {
    super()
  }

  render() {
    return(
      <span className="apps--groupsList">
        {this.props.groups.map((group, i) =>
          <Item key={"group_" + i} group={group} appID={this.props.appID} appName={this.props.appName} />
        )}
      </span>
    )
  }

}

ApplicationItemGroupsList.propTypes = {
  groups: PropTypes.array.isRequired,
  appID: PropTypes.string.isRequired,
  appName: PropTypes.string.isRequired
}

export default ApplicationItemGroupsList
