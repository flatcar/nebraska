import Divider from '@material-ui/core/Divider';
import MuiListItem from '@material-ui/core/ListItem';
import { makeStyles } from '@material-ui/core/styles';
import React from 'react';

const useStyles = makeStyles({
  outterDivider: {
    height: '2px',
  },
});

export default function ListItem(props) {
  const classes = useStyles();

  return (
    <React.Fragment>
      <Divider className={classes.outterDivider} />
      <MuiListItem disableGutters {...props} />
    </React.Fragment>
  );
}
