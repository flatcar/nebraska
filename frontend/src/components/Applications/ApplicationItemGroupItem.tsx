import LayersOutlinedIcon from '@mui/icons-material/LayersOutlined';
import { Box } from '@mui/material';
import Link from '@mui/material/Link';
import React from 'react';
import { useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { Link as RouterLink } from 'react-router';

import { APIContext } from '../../api/API';
import { Group } from '../../api/apiDataTypes';
import ChannelItem from '../Channels/ChannelItem';

function ApplicationItemGroupItem(props: { group: Group; appName: string }) {
  const { group } = props;
  const [totalInstances, setTotalInstances] = React.useState(-1);
  const { t } = useTranslation();
  const API = useContext(APIContext);

  React.useEffect(() => {
    // We use this function without any filter to get the total number of instances
    // in the group.
    API.getInstancesCount(group.application_id, group.id, '1d')
      .then(result => {
        setTotalInstances(result);
      })
      .catch(err => console.error('Error loading total instances in Instances/List', err));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [group]);

  const instanceCountContent = (
    <Box display="flex">
      <LayersOutlinedIcon />
      <Box px={0.5}>{totalInstances}</Box>
      {t('applications|instances')}
    </Box>
  );

  return (
    <>
      <Box display="flex" p={1}>
        <Box width="40%">
          <Link
            sx={{
              fontSize: '1rem',
              color: '#1b5c91',
            }}
            to={{ pathname: `/apps/${props.group.application_id}/groups/${props.group.id}` }}
            component={RouterLink}
            underline="hover"
          >
            {props.group.name}
          </Link>
        </Box>
        <Box display="flex" width="50%">
          {totalInstances > 0 ? (
            <Link
              to={{
                pathname: `/apps/${props.group.application_id}/groups/${props.group.id}/instances`,
                search: 'period=1d',
              }}
              component={RouterLink}
              sx={{
                color: '#1b5c91',
              }}
              underline="hover"
            >
              {instanceCountContent}
            </Link>
          ) : (
            instanceCountContent
          )}
        </Box>
      </Box>
      {group.channel && <ChannelItem channel={group.channel} isAppView />}
    </>
  );
}

export default ApplicationItemGroupItem;
