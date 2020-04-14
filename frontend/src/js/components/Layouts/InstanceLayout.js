import React from 'react';
import { applicationsStore, instancesStore } from '../../stores/Stores';
import Loader from '../Common/Loader';
import SectionHeader from '../Common/SectionHeader';
import Details from '../Instances/Details';

export default function InstanceLayout(props) {
  const {appID, groupID, instanceID} = props.match.params;
  const [application, setApplication] = React.useState(
    applicationsStore
      .getCachedApplication(appID)
  );
  const [group, setGroup] = React.useState(getGroupFromApplication(application));
  const [instance, setInstance] = React.useState(null);

  function onChange() {
    instancesStore.getInstances(appID, groupID, instanceID);
    const apps = applicationsStore.getCachedApplications() || [];
    const app = apps.find(({id}) => id === appID);
    if (app !== application) {
      setApplication(app);
      setGroup(getGroupFromApplication(app));
    }

    onChangeInstances();
  }

  function getGroupFromApplication(app) {
    return app ? app.groups.find(({id}) => id === groupID) : null;
  }

  React.useEffect(() => {
    applicationsStore.addChangeListener(onChange);

    applicationsStore.getApplication(appID);

    return function cleanup() {
      applicationsStore.removeChangeListener(onChange);
      instancesStore.removeChangeListener(onChangeInstances);
    };
  },
  []);

  const applicationName = application ? application.name : '…';
  const groupName = group ? group.name : '…';

  return (
    <React.Fragment>
      <SectionHeader
        title={instanceID}
        breadcrumbs={[
          {
            path: '/apps',
            label: 'Applications'
          },
          {
            path: `/apps/${appID}`,
            label: applicationName
          },
          {
            path: `/apps/${appID}/groups/${groupID}`,
            label: groupName
          },
          {
            path: `/apps/${appID}/groups/${groupID}/instances`,
            label: 'Instances'
          },
        ]}
      />
      { !instance ? <Loader />
        :
      <Details
        application={application}
        group={group}
        instance={instance}
      />
      }
    </React.Fragment>
  );
}
