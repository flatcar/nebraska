import PropTypes from 'prop-types';
import React from "react"
import { Link } from "react-router-dom"

class ApplicationItemGroupItem extends React.Component {

  constructor() {
    super()
  }

  render() {
    const instances_total = this.props.group.instances_stats.total ? "(" + this.props.group.instances_stats.total + ")" : ""

    return(
      <Link to={{pathname: `/apps/${this.props.group.application_id}/groups/${this.props.group.id}`}}>
        <span className="activeLink lighter">
          {this.props.group.name} {instances_total}&nbsp;<i className="fa fa-caret-right"></i>
        </span>
      </Link>
    )
  }

}

ApplicationItemGroupItem.propTypes = {
  group: PropTypes.object.isRequired,
  appName: PropTypes.string.isRequired
}

export default ApplicationItemGroupItem
