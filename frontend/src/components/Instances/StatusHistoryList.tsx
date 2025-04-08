import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import makeStyles from '@mui/styles/makeStyles';
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
function StatusHistoryList(props: { entries: InstanceStatusHistory[] }) {
  const entries = props.entries || [];
  const classes = useStyles();
  const { t } = useTranslation();

  // @todo: Virtualize the table.
  return (
    <Table className={classes.root}>
      <TableHead>
        <TableRow>
          <TableCell>{t('instances|instances_plural')}</TableCell>
          <TableCell>{t('instances|status')}</TableCell>
          <TableCell>{t('instances|message')}</TableCell>
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
