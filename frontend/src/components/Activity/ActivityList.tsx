import MuiList from '@mui/material/List';
import Typography from '@mui/material/Typography';
import makeStyles from '@mui/styles/makeStyles';
import React from 'react';

import { Activity } from '../../api/apiDataTypes';
import { makeLocaleTime } from '../../i18n/dateTime';
import ActivityItem from './ActivityItem';

const useStyles = makeStyles({
  listTitle: {
    fontSize: '1em',
  },
});

export interface ActivityListProps {
  entries?: Activity[];
  timestamp: string;
}

export default function ActivityList(props: ActivityListProps) {
  const classes = useStyles();
  const entries = props.entries ? props.entries : [];

  return (
    <React.Fragment>
      <Typography className={classes.listTitle}>
        {makeLocaleTime(props.timestamp, {
          showTime: false,
          dateFormat: { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' },
        })}
      </Typography>
      <MuiList>
        {entries.map((entry: Activity) => (
          <ActivityItem key={entry.id} entry={entry} />
        ))}
      </MuiList>
    </React.Fragment>
  );
}
