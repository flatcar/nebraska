import Grid from '@material-ui/core/Grid';
import React from 'react';
import { useParams } from 'react-router-dom';
import _ from 'underscore';
import { applicationsStore } from '../../stores/Stores';
import ChannelsList from '../Channels/List';
import SectionHeader from '../Common/SectionHeader';
import GroupsList from '../Groups/List';
import PackagesList from '../Packages/List';

function ApplicationLayout(props: {}) {
  const {appID} = useParams<{appID: string}>();
  const [applications, setApplications] =
      React.useState(applicationsStore.getCachedApplications() || []);

  function onChange() {
    setApplications(applications);
  }

  React.useEffect(() => {
    applicationsStore.addChangeListener(onChange);
    return () => {
      applicationsStore.removeChangeListener(onChange);
    };
  });

  let appName = '';
  const application = _.findWhere(applications, {id: appID});

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
            label: 'Applications'
          }
        ]}
      />
      <Grid
        container
        spacing={1}
        justify="space-between"
      >
        <Grid item xs={8}>
          <GroupsList appID={appID} />
        </Grid>
        <Grid item xs={4}>
          <Grid
            container
            direction="column"
            alignItems="stretch"
            spacing={2}
          >
            <Grid item xs={12}>
              <ChannelsList appID={appID} />
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
