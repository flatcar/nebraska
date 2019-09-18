import React from 'react';
import { applicationsStore, instancesStore } from '../../stores/Stores';
import Loader from '../Common/Loader';
import SectionHeader from '../Common/SectionHeader';
import Details from '../Instances/Details';

export default function InstanceLayout(props) {
  let {appID, groupID, instanceID} = props.match.params;
  let [application, setApplication] = React.useState(applicationsStore.getCachedApplication(appID));
  let [group, setGroup] = React.useState(getGroupFromApplication(application));
  let [instance, setInstance] = React.useState(null);

  function onChange() {
    instancesStore.getInstances(appID, groupID, instanceID);
    let apps = applicationsStore.getCachedApplications() || [];
    let app = apps.find(({id}) => id === appID);
    if (app !== application) {
      setApplication(app);
      setGroup(getGroupFromApplication(app));
    }

    onChangeInstances();
  }

  function getGroupFromApplication(app) {
    return app ? app.groups.find(({id}) => id === groupID) : null;
  }

  function onChangeInstances() {
    let instances = instancesStore.getCachedInstances(appID, groupID) || [];
    let inst = instances.find(({id}) => id === instanceID);
    if (inst !== instance) {
      setInstance(inst);
    }
  }

  React.useEffect(() => {
    applicationsStore.addChangeListener(onChange);
    instancesStore.addChangeListener(onChangeInstances);

    applicationsStore.getApplication(appID);

    return function cleanup() {
      applicationsStore.removeChangeListener(onChange);
      instancesStore.removeChangeListener(onChangeInstances);
    };
  },
  []);

  let applicationName = application ? application.name : '…';
  let groupName = group ? group.name : '…';

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
        }
        ]}
      />
      { !instance ? <Loader />
      : <Details
          application={application}
          group={group}
          instance={instance}
        />
      }
    </React.Fragment>
  );
}
