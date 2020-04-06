import React from 'react';
import { applicationsStore } from '../../stores/Stores';
import Loader from '../Common/Loader';
import SectionHeader from '../Common/SectionHeader';
import List from '../Instances/List';

export default function InstanceLayout(props) {
  const {appID, groupID} = props.match.params;
  const [application, setApplication] =
    React.useState(applicationsStore.getCachedApplication(appID));
  const [group, setGroup] = React.useState(getGroupFromApplication(application));

  function onChange() {
    const apps = applicationsStore.getCachedApplications() || [];
    const app = apps.find(({id}) => id === appID);
    if (app !== application) {
      setApplication(app);
      setGroup(getGroupFromApplication(app));
    }
  }

  function getGroupFromApplication(app) {
    return app ? app.groups.find(({id}) => id === groupID) : null;
  }

  React.useEffect(() => {
    applicationsStore.addChangeListener(onChange);
    applicationsStore.getApplication(appID);

    return function cleanup() {
      applicationsStore.removeChangeListener(onChange);
    };
  },
  []);

  const applicationName = application ? application.name : '…';
  const groupName = group ? group.name : '…';

  return (
    <React.Fragment>
      <SectionHeader
        title="Instances"
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
          }
        ]}
      />
      { group === null ? <Loader />
        :
      <List
        application={application}
        group={group}
      />
      }
    </React.Fragment>
  );
}
