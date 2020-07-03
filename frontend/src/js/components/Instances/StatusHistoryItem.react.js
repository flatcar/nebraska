import { Box } from '@material-ui/core';
import Chip from '@material-ui/core/Chip';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import PropTypes from 'prop-types';
import React from 'react';
import { getInstanceStatus, makeLocaleTime } from '../../constants/helpers';

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
    const status = getInstanceStatus(this.props.entry.status, this.props.entry.version);
    this.setState({status: status});
  }

  render() {
    const time = makeLocaleTime(this.props.entry.created_ts);
    const {className, bgColor, textColor, status} = this.state.status;
    const instanceLabel = className ?
      <Box p={1} bgcolor={bgColor} color={textColor} textAlign="center">
        {status}
      </Box> :
      <div>&nbsp;</div>;

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
