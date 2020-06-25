import { Box, Grid, Link, makeStyles, Typography } from '@material-ui/core';
import React from 'react';
import API from '../../api/API';
import { timeIntervalsDefault } from '../../constants/helpers';

const useStyles = makeStyles(() => ({
  title: {
    fontSize: '1rem'
  }
}));

export default function TimeIntervalLinks(props) {
  const {selectedInterval, intervalChangeHandler} = props;
  const [timeIntervals, setTimeIntervals] = React.useState(timeIntervalsDefault);
  const { appID, groupID } = props;
  const classes = useStyles();

  React.useEffect(() => {
    if (appID && groupID) {
      Promise.all(timeIntervalsDefault.map((timeInterval) => {
        return API.getInstancesCount(appID, groupID, timeInterval.queryValue);
      })).then((results) => {
        const timeIntervalsToUpdate = [...timeIntervals];
        for (let i = 0; i < results.length; i++) {
          if (results[i] === 0) {
            const timeInterval = {...timeIntervalsToUpdate[i]};
            timeInterval.disabled = true;
            timeIntervalsToUpdate[i] = timeInterval;
          }
        }
        setTimeIntervals(timeIntervalsToUpdate);
      });
    }

  }, [appID, groupID]);

  return (
    <Grid container spacing={1}>
      {
        timeIntervals.map((link, index) =>
          <>
            <Grid key={link.queryValue} item>
              <Link underline="none"
                component="button"
                onClick={(e) => intervalChangeHandler(link)}
                color={link.disabled ? 'textSecondary' : link.displayValue === selectedInterval.displayValue ? 'inherit' : 'primary'}
              >
                <Typography className={classes.title} >
                  {link.displayValue}
                </Typography>
              </Link>
            </Grid>
            {(index < timeIntervals.length - 1) && <Grid item><Box color="text.disabled">{'.'}</Box></Grid>}
          </>)
      }
    </Grid>
  );
}

