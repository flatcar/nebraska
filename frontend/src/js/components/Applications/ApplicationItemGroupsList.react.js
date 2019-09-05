import Grid from '@material-ui/core/Grid';
import PropTypes from 'prop-types';
import React from 'react';
import Item from './ApplicationItemGroupItem.react';

function ApplicationItemGroupsList(props) {
  return(
    <Grid container spacing={2}>
      {props.groups.map((group, i) =>
        <Grid item>
          <Item key={"group_" + i} group={group} appID={props.appID} appName={props.appName} />
        </Grid>
      )}
    </Grid>
  );
}

ApplicationItemGroupsList.propTypes = {
  groups: PropTypes.array.isRequired,
  appID: PropTypes.string.isRequired,
  appName: PropTypes.string.isRequired
};

export default ApplicationItemGroupsList;
