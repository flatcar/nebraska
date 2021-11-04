import { Box, Divider } from '@material-ui/core';
import { Fragment } from 'react';
import { Group } from '../../api/apiDataTypes';
import ApplicationItemGroupItem from './ApplicationItemGroupItem';

export default function ApplicationItemGroupsList(props: {
  groups: Group[];
  appID: string;
  appName: string;
}) {
  return (
    <>
      {props.groups.map((group, i) => (
        <Fragment key={group.id}>
          {i > 0 && <Divider variant="fullWidth" />}
          <Box mt={1}>
            <ApplicationItemGroupItem group={group} appName={props.appName} key={'group_' + i} />
          </Box>
        </Fragment>
      ))}
    </>
  );
}
