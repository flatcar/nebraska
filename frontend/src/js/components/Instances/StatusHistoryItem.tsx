import { Box } from '@material-ui/core';
import Chip from '@material-ui/core/Chip';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import React from 'react';
import { ERROR_STATUS_CODE, getErrorAndFlags, getInstanceStatus, makeLocaleTime, prepareErrorMessage } from '../../constants/helpers';

interface StatusHistoryItemProps {
  entry: {
    status: number;
    version?: string | undefined;
    created_ts: string;
    error_code: number;
  };
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
  },
  []);

  function fetchStatusFromStore() {
    const status = getInstanceStatus(props.entry.status, props.entry.version);
    setStatus(status);
  }

  const time = makeLocaleTime(props.entry.created_ts);
  const {className, bgColor, textColor} = status;
  const errorCode = props.entry.error_code;
  let extendedErrorExplanation = '';
  if (props.entry.status === ERROR_STATUS_CODE) {
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
        {status.explanation}
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

export default StatusHistoryItem;
