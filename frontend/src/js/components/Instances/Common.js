import Grid from '@material-ui/core/Grid';
import Typography from '@material-ui/core/Typography';
import { makeStyles } from '@material-ui/styles';
import React from 'react';

const useInstanceCountStyles = makeStyles(theme => ({
  instancesCount: {
    fontSize: '3rem;'
  },
  instancesLabel: {
    color: theme.palette.text.secondary,
    fontSize: '1.5rem;'
  },
}));

export function InstanceCountLabel(props) {
  const classes = useInstanceCountStyles();
  const {countText} = props;

  return (
    <Grid
      container
      alignItems="center"
      justify="center"
      direction="column"
    >
      <Grid item>
        <Typography className={classes.instancesCount}>{countText}</Typography>
      </Grid>
      <Grid item>
        <Typography className={classes.instancesLabel}>Instances</Typography>
      </Grid>
    </Grid>
  );
}
