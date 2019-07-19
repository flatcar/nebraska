import PropTypes from 'prop-types';
import { instancesStore } from "../../stores/Stores"
import React from "react"
import { Label } from "react-bootstrap"
import moment from "moment"
import _ from "underscore"

class StatusHistoryItem extends React.Component {

  constructor(props) {
    super(props)
    this.fetchStatusFromStore = this.fetchStatusFromStore.bind(this)

    this.state = {status: {}}
  }

  componentDidMount() {
    this.fetchStatusFromStore()
  }

  fetchStatusFromStore() {
    let status = instancesStore.getInstanceStatus(this.props.entry.status, this.props.entry.version)
    this.setState({status: status})
  }

  render() {
    let date = moment.utc(this.props.entry.created_ts).local().format("DD/MM/YYYY"),
        time = moment.utc(this.props.entry.created_ts).local().format("hh:mma"),
        instanceLabel = this.state.status.className ? <Label>{this.state.status.status}</Label> : <div>&nbsp;</div>

    return(
      <li>
        <div className="event--date">
          {date}
          <span>{time}</span>
        </div>
        <div>
          {instanceLabel}
        </div>
        <div>
          <p>
            {this.state.status.explanation}
          </p>
        </div>
      </li>
    )
  }

}

StatusHistoryItem.propTypes = {
  entry: PropTypes.object.isRequired
}

export default StatusHistoryItem
