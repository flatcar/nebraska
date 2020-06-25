import { Box } from '@material-ui/core';
import Grid from '@material-ui/core/Grid';
import { makeStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import React from 'react';

const useStyles = makeStyles(theme => ({
  sectionHeader: {
    padding: '1em',
  },
}));

export default function ListHeader(props) {
  const classes = useStyles();
  const actions = props.actions || [];
  return (
    <Grid
      container
      alignItems="flex-start"
      justify="space-between"
      className={classes.sectionHeader}
    >
      {props.title &&
      <Grid item>
        <Box p={0.5}>
          <Typography variant="h4">{props.title}</Typography>
        </Box>
      </Grid>
      }
      {actions &&
              actions.map((action, i) =>
                <Grid item key={i}>
                  {action}
                </Grid>
              )
      }
    </Grid>
  );
}
