import React from 'react';
import { useTranslation } from 'react-i18next';
import { Navigate, useParams } from 'react-router';

import API from '../../../api/API';
import { Application, Instance } from '../../../api/apiDataTypes';
import { applicationsStore } from '../../../stores/Stores';
import { getInstanceStatus } from '../../../utils/helpers';
import Loader from '../../common/Loader';
import SectionHeader from '../../common/SectionHeader';
import Details from '../../Instances/Details';

export default function InstanceLayout() {
  const { appID, groupID, instanceID } = useParams<{
    appID: string;
    groupID: string;
    instanceID: string;
  }>();

  const [application, setApplication] = React.useState(
    applicationsStore().getCachedApplication(appID || '')
  );
  const [instance, setInstance] = React.useState<Instance | null>(null);
  const instanceRequestID = React.useRef(0);
  const { t } = useTranslation();

  const getGroupFromApplication = React.useCallback(
    (app: Application | null) => {
      if (!app) {
        return null;
      }
      const group = app.groups.find(({ id }) => id === groupID);
      return group || null;
    },
    [groupID]
  );

  const group = getGroupFromApplication(application || null);

  const onChange = React.useCallback(() => {
    if (!appID || !groupID || !instanceID) {
      return;
    }

    const requestID = ++instanceRequestID.current;
    API.getInstance(appID, groupID, instanceID).then(instance => {
      if (requestID !== instanceRequestID.current) {
        return;
      }
      instance.statusInfo = getInstanceStatus(
        instance.application.status,
        instance.application.version
      );
      setInstance(instance);
    });
    const apps = applicationsStore().getCachedApplications() || [];
    const app = apps.find(({ id }) => id === appID) || null;
    setApplication(currentApplication => (app === currentApplication ? currentApplication : app));
  }, [appID, groupID, instanceID]);

  React.useLayoutEffect(() => {
    setInstance(null);
    return () => {
      instanceRequestID.current += 1;
    };
  }, [appID, groupID, instanceID]);

  React.useEffect(() => {
    if (!appID) {
      return;
    }
    applicationsStore().addChangeListener(onChange);

    applicationsStore().getApplication(appID);

    return function cleanup() {
      applicationsStore().removeChangeListener(onChange);
    };
  }, [appID, onChange]);

  const applicationName = application ? application.name : '…';
  const groupName = group ? group.name : '…';
  const instanceName = (instance && instance.alias) || instanceID;

  const searchParams = new URLSearchParams(window.location.search).toString();

  if (!appID || !instanceName) {
    return <Navigate to="/404" replace />;
  }

  return (
    <React.Fragment>
      <SectionHeader
        title={instanceName}
        breadcrumbs={[
          {
            path: '/apps',
            label: t('layouts|applications'),
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
            label: t('layouts|instances'),
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
