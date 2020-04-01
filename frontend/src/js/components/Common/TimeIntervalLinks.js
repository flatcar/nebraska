import { Grid, Link, makeStyles, Typography } from '@material-ui/core';
import React from 'react';
import timeIntervals from '../../constants/timeInterval';

const useStyles = makeStyles(theme => ({
  title:{
    fontSize:'1rem'
  }
}));

export default function TimeIntervalLinks(props){
  const [activeLink, setActiveLink]=React.useState(timeIntervals[0]);
  const classes=useStyles();
  function handleIntervalSelect(link){
    setActiveLink(link);
    props.intervalChangeHandler(link);
  }
  return (
    <Grid container spacing={1}>
      {
        timeIntervals.map((link)=> <Grid key={link} item onClick={(e)=>handleIntervalSelect(link)}>
          <Link underline="none" component="button" color={link===activeLink?'inherit':'primary'}>
            <Typography className={classes.title}>
              {link}
            </Typography>
          </Link></Grid>)
      }

    </Grid>
  );
}

