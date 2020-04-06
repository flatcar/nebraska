import Chip from '@material-ui/core/Chip';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import PropTypes from 'prop-types';
import React from 'react';
import { makeLocaleTime } from '../../constants/helpers';
import { instancesStore } from '../../stores/Stores';

class StatusHistoryItem extends React.Component {

  constructor(props) {
    super(props);
    this.fetchStatusFromStore = this.fetchStatusFromStore.bind(this);

    this.state = {status: {}};
  }

  componentDidMount() {
    this.fetchStatusFromStore();
  }

  fetchStatusFromStore() {
    const status = instancesStore
      .getInstanceStatus(this.props.entry.status, this.props.entry.version);
    this.setState({status: status});
  }

  render() {
    const time = makeLocaleTime(this.props.entry.created_ts);
    const instanceLabel = this.state.status.className ? <Chip size='small' label={this.state.status.status} /> : <div>&nbsp;</div>;

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
};

export default StatusHistoryItem;
