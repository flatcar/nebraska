import { Grid, Link, makeStyles, Typography } from '@material-ui/core';
import React from 'react';
import { defaultTimeInterval, timeIntervals } from '../../constants/helpers';

const useStyles = makeStyles(() => ({
  title: {
    fontSize: '1rem'
  }
}));

export default function TimeIntervalLinks(props) {
  const [activeLink, setActiveLink] = React.useState(defaultTimeInterval.displayValue);
  const classes = useStyles();
  function handleIntervalSelect(link) {
    setActiveLink(link.displayValue);
    props.intervalChangeHandler(link);
  }
  return (
    <Grid container spacing={1}>
      {
        timeIntervals.map((link) =>
          <Grid key={link.queryValue}
            item
            onClick={(e) => handleIntervalSelect(link)}
          >
            <Link underline="none" component="button" color={link.displayValue === activeLink ? 'inherit' : 'primary'}>
              <Typography className={classes.title}>
                {link.displayValue}
              </Typography>
            </Link>
          </Grid>)
      }
    </Grid>
  );
}

