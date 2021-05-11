import { Box } from '@material-ui/core';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import React from 'react';
import { InstanceStatusHistory } from '../../api/apiDataTypes';
import {
  ERROR_STATUS_CODE,
  getErrorAndFlags,
  getInstanceStatus,
  makeLocaleTime,
  prepareErrorMessage
} from '../../utils/helpers';

interface StatusHistoryItemProps {
  entry: InstanceStatusHistory;
}

function StatusHistoryItem(props: StatusHistoryItemProps) {
  const [status, setStatus] = React.useState<{
    type?: string;
    className?: string;
    spinning?: boolean;
    icon?: string;
    description?: string;
    status?: string;
    explanation?: string;
    textColor?: string | undefined;
    bgColor?: string | undefined;
  }>({});
  React.useEffect(() => {
    fetchStatusFromStore();
  }, []);

  function fetchStatusFromStore() {
    const status = getInstanceStatus(props.entry.status, props.entry.version);
    setStatus(status);
  }

  const time = makeLocaleTime(props.entry.created_ts);
  const { className, bgColor, textColor, status: statusString } = status;
  const errorCode = props.entry.error_code;
  let extendedErrorExplanation = '';
  if (props.entry.status === ERROR_STATUS_CODE && errorCode !== null) {
    const [errorMessages, flags] = getErrorAndFlags(parseInt(errorCode));
    extendedErrorExplanation = prepareErrorMessage(errorMessages, flags);
  }
  const instanceLabel = className ? (
    <Box p={1} bgcolor={bgColor} color={textColor} textAlign="center">
      {statusString}
    </Box>
  ) : (
    <div>&nbsp;</div>
  );

  return (
    <TableRow>
      <TableCell>{time}</TableCell>
      <TableCell>{instanceLabel}</TableCell>
      <TableCell>
        {status.explanation}
        {extendedErrorExplanation && (
          <>
            {':'}
            <Box>{extendedErrorExplanation}</Box>
          </>
        )}
      </TableCell>
    </TableRow>
  );
}

export default StatusHistoryItem;
