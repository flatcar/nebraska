import React from 'react';
import { useTranslation } from 'react-i18next';
import { useParams } from 'react-router-dom';
import API from '../../api/API';
import { Application, Instance } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { getInstanceStatus } from '../../utils/helpers';
import Loader from '../Common/Loader';
import SectionHeader from '../Common/SectionHeader';
import Details from '../Instances/Details';

export default function InstanceLayout() {
  const { appID, groupID, instanceID } = useParams<{
    appID: string;
    groupID: string;
    instanceID: string;
  }>();
  const [application, setApplication] = React.useState(
    applicationsStore.getCachedApplication(appID)
  );
  const [group, setGroup] = React.useState(getGroupFromApplication(application || null));
  const [instance, setInstance] = React.useState<Instance | null>(null);
  const { t } = useTranslation();

  function onChange() {
    API.getInstance(appID, groupID, instanceID).then(instance => {
      instance.statusInfo = getInstanceStatus(
        instance.application.status,
        instance.application.version
      );
      setInstance(instance);
    });
    const apps = applicationsStore.getCachedApplications() || [];
    const app = apps.find(({ id }) => id === appID) || null;
    if (app !== application) {
      setApplication(app);
      setGroup(getGroupFromApplication(app));
    }
  }

  function getGroupFromApplication(app: Application | null) {
    if (!app) {
      return null;
    }
    const group = app.groups.find(({ id }) => id === groupID);
    return group || null;
  }

  React.useEffect(() => {
    applicationsStore.addChangeListener(onChange);

    applicationsStore.getApplication(appID);

    return function cleanup() {
      applicationsStore.removeChangeListener(onChange);
    };
  }, []);

  const applicationName = application ? application.name : '…';
  const groupName = group ? group.name : '…';
  const instanceName = (instance && instance.alias) || instanceID;

  const searchParams = new URLSearchParams(window.location.search).toString();
  return (
    <React.Fragment>
      <SectionHeader
        title={instanceName}
        breadcrumbs={[
          {
            path: '/apps',
            label: t('layouts|Applications'),
          },
          {
            path: `/apps/${appID}`,
            label: applicationName,
          },
          {
            path: `/apps/${appID}/groups/${groupID}`,
            label: groupName,
          },
          {
            path: `/apps/${appID}/groups/${groupID}/instances?${searchParams}`,
            label: t('layouts|Instances'),
          },
        ]}
      />
      {!instance ? (
        <Loader />
      ) : (
        <Details
          application={application!}
          group={group!}
          instance={instance}
          onInstanceUpdated={() => onChange()}
        />
      )}
    </React.Fragment>
  );
}
