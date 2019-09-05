import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import PropTypes from 'prop-types';
import React from 'react';
import StatusHistoryItem from './StatusHistoryItem.react';

function StatusHistoryList(props) {
  let entries = props.entries || [];

  // @todo: Virtualize the table.
  return(
    <Table>
      <TableHead>
        <TableCell>Instances</TableCell>
        <TableCell>Status</TableCell>
        <TableCell>Message</TableCell>
      </TableHead>
      <TableBody>
        {entries.map((entry, i) =>
          <StatusHistoryItem key={"statusHistory_" + i} entry={entry} />
        )}
      </TableBody>
    </Table>
  );
}

StatusHistoryList.propTypes = {
  entries: PropTypes.array.isRequired
}

export default StatusHistoryList
