import React from 'react';
import { useParams } from 'react-router-dom';
import _ from 'underscore';
import { Channel, Group } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import SectionHeader from '../Common/SectionHeader';
import EditDialog from '../Groups/EditDialog';
import GroupExtended from '../Groups/ItemExtended';

function GroupLayout(props: {}) {
  const {appID, groupID} = useParams<{appID: string; groupID: string}>();
  const [applications, setApplications] =
      React.useState(applicationsStore.getCachedApplications() || []);
  const [updateGroupModalVisible, setUpdateGroupModalVisible] = React.useState(false);

  React.useEffect(() => {
    applicationsStore.getApplication(appID);
    applicationsStore.addChangeListener(onChange);
    return () => {
      applicationsStore.removeChangeListener(onChange);
    };
  },
  []);

  function onChange() {
    setApplications(applicationsStore.getCachedApplications() || []);
  }

  function openUpdateGroupModal() {
    setUpdateGroupModalVisible(true);
  }

  function closeUpdateGroupModal() {
    setUpdateGroupModalVisible(false);
  }

  let appName = '';
  let groupName = '';

  const application = _.findWhere(applications, {id: appID});
  let groups: Group[] = [];
  let channels: Channel[] = [];

  if (application) {
    appName = application.name;
    groups = application.groups;
    channels = application.channels ? application.channels : [];
    const group = _.findWhere(application.groups, {id: groupID});
    if (group) {
      groupName = group.name;
    }
  }

  const groupToUpdate = _.findWhere(groups, {id: groupID});

  return (
    <div>
      <SectionHeader
        title={groupName}
        breadcrumbs={[
          {
            path: '/apps',
            label: 'Applications'
          },
          {
            path: `/apps/${appID}`,
            label: appName
          }
        ]}
      />
      <GroupExtended
        appID={appID}
        groupID={groupID}
        handleUpdateGroup={openUpdateGroupModal}
      />
      <EditDialog
        data={{group: groupToUpdate, channels: channels}}
        show={updateGroupModalVisible}
        onHide={closeUpdateGroupModal}
      />
    </div>
  );
}

export default GroupLayout;
