import PropTypes from 'prop-types';
import React from "react"
import StatusHistoryList from "./StatusHistoryList.react"
import _ from "underscore"

class StatusHistoryContainer extends React.Component {

  constructor(props) {
    super(props)
  }

  render() {
    let entries = "",
        additionalStyle = ""

    if (_.isEmpty(this.props.instance.statusHistory)) {
      entries = <div className="emptyBox">This instance hasn’t reported any events yet in the context of this application/group.</div>
      additionalStyle = " coreRollerTable-detail--empty"
    } else {
      entries = <StatusHistoryList entries={this.props.instance.statusHistory} />
    }

    return(
      <div className={"coreRollerTable-detail" + additionalStyle + this.props.active} id={"detail-" + this.props.key}>
        <div className="coreRollerTable-detailContent">
          {entries}
        </div>
      </div>
    )
  }

}

StatusHistoryContainer.propTypes = {
  key: PropTypes.string.isRequired,
  active: PropTypes.array.isRequired,
  instance: PropTypes.object.isRequired
}

export default StatusHistoryContainer
