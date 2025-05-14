import { Box, Grid, Link, Typography } from '@mui/material';
import { styled } from '@mui/material/styles';
import React from 'react';

import API from '../../../api/API';
import { timeIntervalsDefault } from '../../../utils/helpers';

const PREFIX = 'TimeIntervalLinks';

const classes = {
  title: `${PREFIX}-title`,
};

const StyledGrid = styled(Grid)(() => ({
  [`& .${classes.title}`]: {
    fontSize: '1rem',
  },
}));

interface TimeIntervalLinksProps {
  selectedInterval: string;
  appID?: string;
  groupID?: string;
  intervalChangeHandler: (value: any) => any;
}

export default function TimeIntervalLinks(props: TimeIntervalLinksProps) {
  const { selectedInterval, intervalChangeHandler } = props;
  const [timeIntervals, setTimeIntervals] = React.useState(timeIntervalsDefault);
  const { appID, groupID } = props;

  React.useEffect(() => {
    if (appID && groupID) {
      Promise.all(
        timeIntervalsDefault.map(timeInterval => {
          return API.getInstancesCount(appID, groupID, timeInterval.queryValue);
        })
      ).then(results => {
        const timeIntervalsToUpdate = [...timeIntervals];
        for (let i = 0; i < results.length; i++) {
          if (results[i] === 0) {
            const timeInterval = { ...timeIntervalsToUpdate[i] };
            timeInterval.disabled = true;
            timeIntervalsToUpdate[i] = timeInterval;
          }
        }
        setTimeIntervals(timeIntervalsToUpdate);
      });
    }
  }, [appID, groupID]);

  return (
    <StyledGrid container spacing={1}>
      {timeIntervals.map((link, index) => (
        <React.Fragment key={link.queryValue}>
          <Grid>
            <Link
              underline="none"
              component="button"
              onClick={() => intervalChangeHandler(link)}
              color={
                link.disabled
                  ? 'textSecondary'
                  : link.queryValue === selectedInterval
                    ? 'inherit'
                    : 'primary'
              }
              style={{
                color: !link.disabled && link.queryValue !== selectedInterval ? '#1b5c91' : '',
              }}
            >
              <Typography className={classes.title}>{link.displayValue}</Typography>
            </Link>
          </Grid>
          {index < timeIntervals.length - 1 && (
            <Grid>
              <Box color="text.disabled">{'.'}</Box>
            </Grid>
          )}
        </React.Fragment>
      ))}
    </StyledGrid>
  );
}
