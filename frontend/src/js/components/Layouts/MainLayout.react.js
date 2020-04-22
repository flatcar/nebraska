import Grid from '@material-ui/core/Grid';
import { useTheme } from '@material-ui/core/styles';
import useMediaQuery from '@material-ui/core/useMediaQuery';
import PropTypes from 'prop-types';
import React from 'react';
import ActivityContainer from '../Activity/Container.react';
import ApplicationsList from '../Applications/List.react';

function MainLayout() {
  const theme = useTheme();
  const isSmall = useMediaQuery(theme.breakpoints.down('sm'));
  return (
    <Grid
      container
      spacing={isSmall ? 0 : 1}
      justify="center"
      alignItems="flex-start"
    >
      <Grid item xs={12} md={7} lg={8}>
        <ApplicationsList />
      </Grid>
      <Grid item xs={12} md={5} lg={4}>
        <ActivityContainer />
      </Grid>
    </Grid>
  );
}

export default MainLayout;
