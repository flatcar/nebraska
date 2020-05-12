import { Grid, Link, makeStyles, Typography } from '@material-ui/core';
import React from 'react';
import { timeIntervals } from '../../constants/helpers';

const useStyles = makeStyles(() => ({
  title: {
    fontSize: '1rem'
  }
}));

export default function TimeIntervalLinks(props) {
  const {selectedInterval, intervalChangeHandler} = props;
  const classes = useStyles();
  return (
    <Grid container spacing={1}>
      {
        timeIntervals.map((link) =>
          <Grid key={link.queryValue}
            item
            onClick={(e) => intervalChangeHandler(link)}
          >
            <Link underline="none" component="button" color={link.displayValue === selectedInterval.displayValue ? 'inherit' : 'primary'}>
              <Typography className={classes.title}>
                {link.displayValue}
              </Typography>
            </Link>
          </Grid>)
      }
    </Grid>
  );
}

