import React, { PropTypes } from "react"
import { Link } from "react-router-dom"

class ApplicationItemGroupItem extends React.Component {

  constructor() {
    super()
  }

  static PropTypes: {
    group: React.PropTypes.object.isRequired,
    appName: React.PropTypes.string.isRequired
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

export default ApplicationItemGroupItem
