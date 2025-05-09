import { Box } from '@mui/material';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import makeStyles from '@mui/styles/makeStyles';
import React from 'react';

const useStyles = makeStyles({
  sectionHeader: {
    padding: '1em',
  },
});

export default function ListHeader(props: { title: string; actions?: React.ReactElement[] }) {
  const classes = useStyles();
  const actions = props.actions || [];
  return (
    <Grid
      container
      alignItems="flex-start"
      justifyContent="space-between"
      className={classes.sectionHeader}
    >
      {props.title && (
        <Grid>
          <Box p={0.5}>
            <Typography variant="h1">{props.title}</Typography>
          </Box>
        </Grid>
      )}
      {actions && actions.map((action, i: number) => <Grid key={i}>{action}</Grid>)}
    </Grid>
  );
}
