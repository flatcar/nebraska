import { Box } from '@material-ui/core';
import Chip from '@material-ui/core/Chip';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import PropTypes from 'prop-types';
import React from 'react';
import { ERROR_STATUS_CODE, getErrorAndFlags, getInstanceStatus, makeLocaleTime, prepareErrorMessage } from '../../constants/helpers';

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
    const {className, bgColor, textColor} = this.state.status;
    const {status} = this.state.status;
    const errorCode = this.props.entry.error_code;
    let extendedErrorExplanation = '';
    if (this.props.entry.status === ERROR_STATUS_CODE) {
      const [errorMessages, flags] = getErrorAndFlags(errorCode);
      extendedErrorExplanation = prepareErrorMessage(errorMessages, flags);
    }
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
          {
            extendedErrorExplanation &&
            <>
              {':'}
              <Box>
                {extendedErrorExplanation}
              </Box>
            </>
          }
        </TableCell>
      </TableRow>
    );
  }
}

StatusHistoryItem.propTypes = {
  entry: PropTypes.object.isRequired
};

export default StatusHistoryItem;
