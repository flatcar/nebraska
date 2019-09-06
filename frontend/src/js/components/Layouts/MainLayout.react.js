import Grid from '@material-ui/core/Grid';
import PropTypes from 'prop-types';
import React from 'react';
import ActivityContainer from '../Activity/Container.react';
import ApplicationsList from '../Applications/List.react';

function MainLayout() {
  return(
    <Grid
      container
      spacing={1}
      justify="center"
      alignItems="flex-start">
      <Grid item xs={8}>
        <ApplicationsList />
      </Grid>
      <Grid item xs={4}>
        <ActivityContainer />
      </Grid>
    </Grid>
  );
}

MainLayout.propTypes = {
  stores: PropTypes.object.isRequired
}

export default MainLayout
