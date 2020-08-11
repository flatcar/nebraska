import { Box, Divider } from '@material-ui/core';
import PropTypes from 'prop-types';
import React from 'react';
import Item from './ApplicationItemGroupItem.react';

function ApplicationItemGroupsList(props) {
  return props.groups.map((group, i) => <>
    {i > 0 && <Divider variant="fullWidth"/>}
    <Box mt={1}>
      <Item group={group} appID={props.appID} appName={props.appName} key={'group_' + i}/>
    </Box>
  </>
  );
}

ApplicationItemGroupsList.propTypes = {
  groups: PropTypes.array.isRequired,
  appID: PropTypes.string.isRequired,
  appName: PropTypes.string.isRequired
};

export default ApplicationItemGroupsList;
