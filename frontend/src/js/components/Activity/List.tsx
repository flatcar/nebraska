import MuiList from '@material-ui/core/List';
import { makeStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import React from 'react';
import { Activity } from '../../api/apiDataTypes';
import { makeLocaleTime } from '../../utils/helpers';
import Item from './Item';

const useStyles = makeStyles(theme => ({
  listTitle: {
    fontSize: '1em',
  },
}));

function List(props: { entries?: Activity[]; timestamp: string }) {
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
        {entries.map((entry: Activity, i: number) => (
          <Item key={i} entry={entry} />
        ))}
      </MuiList>
    </React.Fragment>
  );
}

export default List;
