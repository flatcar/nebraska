import React from 'react';
import { useTranslation } from 'react-i18next';
import { useParams } from 'react-router-dom';

import { Application, Group } from '../../../api/apiDataTypes';
import { applicationsStore } from '../../../stores/Stores';
import Loader from '../../common/Loader';
import SectionHeader from '../../common/SectionHeader';
import List from '../../Instances/List';

export default function InstanceListLayout() {
  const { appID, groupID } = useParams<{ appID: string; groupID: string }>();
  const [application, setApplication] = React.useState(
    applicationsStore().getCachedApplication(appID)
  );
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

  const [group, setGroup] = React.useState<Group | null>(getGroupFromApplication(application));
  const { t } = useTranslation();

  const onChange = React.useCallback(() => {
    const apps = applicationsStore().getCachedApplications() || [];
    const app = apps.find(({ id }) => id === appID) || null;
    if (app !== application) {
      setApplication(app);
      setGroup(getGroupFromApplication(app));
    }
  }, [appID, application, getGroupFromApplication]);

  React.useEffect(() => {
    applicationsStore().addChangeListener(onChange);
    applicationsStore().getApplication(appID);

    return function cleanup() {
      applicationsStore().removeChangeListener(onChange);
    };
  }, [appID, onChange]);

  const applicationName = application ? application.name : '…';
  const groupName = group ? group.name : '…';

  return (
    <React.Fragment>
      <SectionHeader
        title={t('layouts|instances')}
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
        ]}
      />
      {!group || !application ? <Loader /> : <List application={application} group={group} />}
    </React.Fragment>
  );
}
