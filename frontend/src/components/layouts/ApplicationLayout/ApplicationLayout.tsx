import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Stack from '@mui/material/Stack';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { Navigate, useParams } from 'react-router';
import _ from 'underscore';

import { applicationsStore } from '../../../stores/Stores';
import ChannelList from '../../Channels/ChannelList';
import SectionHeader from '../../common/SectionHeader/SectionHeader';
import GroupList from '../../Groups/GroupList';
import PackagesList from '../../Packages/List';

function ApplicationLayout() {
  const { appID } = useParams<{ appID: string }>();
  const [applications, setApplications] = React.useState(
    applicationsStore().getCachedApplications() || []
  );
  const { t } = useTranslation();

  function onChange() {
    setApplications(applications);
  }

  React.useEffect(() => {
    applicationsStore().addChangeListener(onChange);
    return () => {
      applicationsStore().removeChangeListener(onChange);
    };
  });

  let appName = '';
  const application = _.findWhere(applications, { id: appID });

  if (application) {
    appName = application.name;
  }

  if (!appID) {
    return <Navigate to="/404" replace />;
  }

  return (
    <div>
      <SectionHeader
        title={appName}
        breadcrumbs={[
          {
            path: '/apps',
            label: t('layouts|applications'),
          },
        ]}
      />
      <Grid
        container
        spacing={1}
        sx={{
          justifyContent: 'space-between',
        }}
      >
        <Grid
          size={{
            xs: 12,
            sm: 8,
          }}
        >
          <GroupList appID={appID} />
        </Grid>
        <Grid
          size={{
            xs: 12,
            sm: 4,
          }}
        >
          <Stack
            direction="column"
            spacing={2}
            sx={{
              alignItems: 'stretch',
            }}
          >
            <Box>
              <ChannelList appID={appID} />
            </Box>
            <Box>
              <PackagesList appID={appID} />
            </Box>
          </Stack>
        </Grid>
      </Grid>
    </div>
  );
}

export default ApplicationLayout;
