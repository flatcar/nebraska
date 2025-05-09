import Grid from '@mui/material/Grid';

import ActivityContainer from '../../Activity/ActivityContainer';
import ApplicationList from '../../Applications/ApplicationList';

function MainLayout() {
  return (
    <Grid container spacing={2} justifyContent="center" alignItems="flex-start">
      <Grid
        size={{
          xs: 12,
          sm: 8,
        }}
      >
        <ApplicationList />
      </Grid>
      <Grid
        size={{
          xs: 12,
          sm: 4,
        }}
      >
        <ActivityContainer />
      </Grid>
    </Grid>
  );
}

export default MainLayout;
