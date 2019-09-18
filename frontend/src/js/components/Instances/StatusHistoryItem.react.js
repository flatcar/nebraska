import Chip from '@material-ui/core/Chip';
import { makeLocaleTime } from '../../constants/helpers';
import PropTypes from 'prop-types';
import React from 'react';
import { instancesStore } from '../../stores/Stores';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';

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
    let time = makeLocaleTime(this.props.entry.created_ts),
        instanceLabel = this.state.status.className ? <Chip size='small' label={this.state.status.status} /> : <div>&nbsp;</div>

    return (
      <TableRow>
        <TableCell>
          {time}
        </TableCell>
        <TableCell>
          {instanceLabel}
        </TableCell>
        <TableCell>
          {this.state.status.explanation}
        </TableCell>
      </TableRow>
    );
  }
}

StatusHistoryItem.propTypes = {
  entry: PropTypes.object.isRequired
}

export default StatusHistoryItem
