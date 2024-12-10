import Grid from '@material-ui/core/Grid';
import ActivityContainer from '../../Activity/ActivityContainer';
import ApplicationList from '../../Applications/ApplicationList';

function MainLayout() {
  return (
    <Grid container spacing={2} justifyContent="center" alignItems="flex-start">
      <Grid item xs={12} sm={8}>
        <ApplicationList />
      </Grid>
      <Grid item xs={12} sm={4}>
        <ActivityContainer />
      </Grid>
    </Grid>
  );
}

export default MainLayout;
