import { Box, Divider } from '@material-ui/core';
import React from 'react';
import { Group } from '../../api/apiDataTypes';
import Item from './ApplicationItemGroupItem';

function ApplicationItemGroupsList(props: {
  groups: Group[];
  appID: string;
  appName: string;
}) {
  return (<>
    {props.groups.map((group, i) => <React.Fragment key={group.id}>
      {i > 0 && <Divider variant="fullWidth"/>}
      <Box mt={1}>
        <Item group={group} appName={props.appName} key={'group_' + i}/>
      </Box>
    </React.Fragment>
    )
    }
  </>
  );
}

export default ApplicationItemGroupsList;
