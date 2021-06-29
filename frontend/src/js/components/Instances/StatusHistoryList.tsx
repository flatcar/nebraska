import { makeStyles } from '@material-ui/core';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { InstanceStatusHistory } from '../../api/apiDataTypes';
import StatusHistoryItem from './StatusHistoryItem';

const useStyles = makeStyles({
  root: {
    '& .MuiTableCell-root': {
      borderBottom: 'none',
    },
  },
});
function StatusHistoryList(props: {
  entries: InstanceStatusHistory[];
}) {
  const entries = props.entries || [];
  const classes = useStyles();
  const { t } = useTranslation();

  // @todo: Virtualize the table.
  return (
    <Table className={classes.root}>
      <TableHead>
        <TableRow>
          <TableCell>{t('instances|Instances')}</TableCell>
          <TableCell>{t('instances|Status')}</TableCell>
          <TableCell>{t('instances|Message')}</TableCell>
        </TableRow>
      </TableHead>
      <TableBody>
        {entries.map((entry, i) => (
          <StatusHistoryItem key={'statusHistory_' + i} entry={entry} />
        ))}
      </TableBody>
    </Table>
  );
}

export default StatusHistoryList;
