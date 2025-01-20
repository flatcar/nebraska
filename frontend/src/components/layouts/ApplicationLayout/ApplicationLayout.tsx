import Grid from '@mui/material/Grid';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { useParams } from 'react-router-dom';
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

  return (
    <div>
      <SectionHeader
        title={appName}
        breadcrumbs={[
          {
            path: '/apps',
            label: t('layouts|Applications'),
          },
        ]}
      />
      <Grid container spacing={1} justifyContent="space-between">
        <Grid item xs={12} sm={8}>
          <GroupList appID={appID} />
        </Grid>
        <Grid item xs={12} sm={4}>
          <Grid container direction="column" alignItems="stretch" spacing={2}>
            <Grid item xs={12}>
              <ChannelList appID={appID} />
            </Grid>
            <Grid item xs={12}>
              <PackagesList appID={appID} />
            </Grid>
          </Grid>
        </Grid>
      </Grid>
    </div>
  );
}

export default ApplicationLayout;
