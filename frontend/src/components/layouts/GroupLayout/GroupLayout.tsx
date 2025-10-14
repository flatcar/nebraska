import React from 'react';
import { useTranslation } from 'react-i18next';
import { Navigate, useParams } from 'react-router';
import _ from 'underscore';

import { Channel, Group } from '../../../api/apiDataTypes';
import { applicationsStore } from '../../../stores/Stores';
import SectionHeader from '../../common/SectionHeader';
import GroupEditDialog from '../../Groups/GroupEditDialog/GroupEditDialog';
import GroupItemExtended from '../../Groups/GroupItemExtended';

function GroupLayout() {
  const { appID, groupID } = useParams<{ appID: string; groupID: string }>();
  const [applications, setApplications] = React.useState(
    applicationsStore().getCachedApplications() || []
  );
  const [updateGroupModalVisible, setUpdateGroupModalVisible] = React.useState(false);
  const { t } = useTranslation();

  function onChange() {
    setApplications(applicationsStore().getCachedApplications() || []);
  }

  React.useEffect(() => {
    if (!appID) {
      return;
    }
    applicationsStore().getApplication(appID);
    applicationsStore().addChangeListener(onChange);
    return () => {
      applicationsStore().removeChangeListener(onChange);
    };
  }, [appID]);

  function openUpdateGroupModal() {
    setUpdateGroupModalVisible(true);
  }

  function closeUpdateGroupModal() {
    setUpdateGroupModalVisible(false);
  }

  let appName = '';
  let groupName = '';

  const application = _.findWhere(applications, { id: appID });
  let groups: Group[] = [];
  let channels: Channel[] = [];

  if (application) {
    appName = application.name;
    groups = application.groups;
    channels = application.channels ? application.channels : [];
    const group = _.findWhere(application.groups, { id: groupID });
    if (group) {
      groupName = group.name;
    }
  }

  const groupToUpdate = _.findWhere(groups, { id: groupID });

  if (!appID || !groupID) {
    return <Navigate to="/404" replace />;
  }

  return (
    <div>
      <SectionHeader
        title={groupName}
        breadcrumbs={[
          {
            path: '/apps',
            label: t('layouts|applications'),
          },
          {
            path: `/apps/${appID}`,
            label: appName,
          },
        ]}
      />
      <GroupItemExtended appID={appID} groupID={groupID} handleUpdateGroup={openUpdateGroupModal} />
      <GroupEditDialog
        data={{ group: groupToUpdate, channels: channels }}
        show={updateGroupModalVisible}
        onHide={closeUpdateGroupModal}
      />
    </div>
  );
}

export default GroupLayout;
