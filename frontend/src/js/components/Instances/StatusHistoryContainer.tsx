import { makeStyles } from '@material-ui/core/styles';
import React from 'react';
import _ from 'underscore';
import Empty from '../Common/EmptyContent';
import StatusHistoryList from './StatusHistoryList';

const useStyles = makeStyles({
  historyBox: {
    paddingLeft: '2em',
    paddingRight: '2em',
    maxHeight: '400px',
    overflow: 'auto',
  },
});

function StatusHistoryContainer(props: {
  statusHistory: {
    status: number;
    version?: string | undefined;
    created_ts: string;
    error_code: number;
  }[];
}) {
  const classes = useStyles();
  let entries: React.ReactElement;

  if (_.isEmpty(props.statusHistory)) {
    entries = (
      <Empty>
        This instance hasn’t reported any events yet in the context of this application/group.
      </Empty>
    );
  } else {
    entries = <StatusHistoryList entries={props.statusHistory} />;
  }

  return <div className={classes.historyBox}>{entries}</div>;
}

export default StatusHistoryContainer;
