import PropTypes from 'prop-types';
import React from "react"
import ApplicationsList from "../Applications/List.react"
import ActivityContainer from "../Activity/Container.react"
import { makeStyles } from '@material-ui/core/styles';
import Grid from '@material-ui/core/Grid';

const useStyles = makeStyles(theme => ({
  root: {
    flexGrow: 1
  },
}));

function MainGrid() {
  const classes = useStyles();

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

class MainLayout extends React.Component {

  constructor() {
    super()
  }

  render() {
    return(<MainGrid />);
  }

}

MainLayout.propTypes = {
  stores: PropTypes.object.isRequired
}

export default MainLayout
