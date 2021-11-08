import MuiList from '@material-ui/core/List';
import Paper from '@material-ui/core/Paper';
import { makeStyles } from '@material-ui/core/styles';
import React from 'react';
import { Trans } from 'react-i18next';
import _ from 'underscore';
import { Channel, Group } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import Empty from '../common/EmptyContent';
import ListHeader from '../common/ListHeader';
import Loader from '../common/Loader/Loader';
import ModalButton from '../common/ModalButton';
import GroupEditDialog from './GroupEditDialog';
import GroupItem from './GroupItem';

const useStyles = makeStyles(() => ({
  root: {
    '& > hr:first-child': {
      display: 'none',
    },
  },
}));

export interface GroupListProps {
  appID: string;
}

function GroupList({ appID }: GroupListProps) {
  const classes = useStyles();
  const [application, setApplication] = React.useState(
    applicationsStore().getCachedApplication(appID)
  );
  const [updateGroupModalVisible, setUpdateGroupModalVisible] = React.useState(false);
  const [updateGroupIDModal, setUpdateGroupIDModal] = React.useState<string | null>(null);

  function closeUpdateGroupModal() {
    setUpdateGroupModalVisible(false);
  }

  function openUpdateGroupModal(appID: string, groupID: string) {
    setUpdateGroupModalVisible(true);
    setUpdateGroupIDModal(groupID);
  }

  React.useEffect(() => {
    applicationsStore().addChangeListener(onChange);
    return () => {
      applicationsStore().removeChangeListener(onChange);
    };
  }, []);

  function onChange() {
    setApplication(applicationsStore().getCachedApplication(appID));
  }

  let channels: Channel[] = [];
  let groups: Group[] = [];
  let entries: React.ReactNode = '';

  if (application) {
    groups = application.groups ? application.groups : [];
    channels = application.channels ? application.channels : [];

    if (_.isEmpty(groups)) {
      entries = (
        <Empty>
          <Trans ns="Groups">
            There are no groups for this application yet.
            <br />
            <br />
            Groups help you control how you want to distribute updates to a specific set of
            instances.
          </Trans>
        </Empty>
      );
    } else {
      entries = _.map(groups, group => {
        return (
          <GroupItem
            key={'groupID_' + group.id}
            group={group}
            handleUpdateGroup={openUpdateGroupModal}
          />
        );
      });
    }
  } else {
    entries = <Loader />;
  }

  const groupToUpdate =
    !_.isEmpty(groups) && updateGroupIDModal
      ? _.findWhere(groups, { id: updateGroupIDModal })
      : null;

  return (
    <>
      <ListHeader
        title="Groups"
        actions={[
          <ModalButton
            icon="plus"
            modalToOpen="AddGroupModal"
            data={{
              channels: channels,
              appID: appID,
            }}
          />,
        ]}
      />
      <Paper>
        <MuiList className={classes.root}>{entries}</MuiList>
        {groupToUpdate && (
          <GroupEditDialog
            data={{ group: groupToUpdate, channels: channels, appID: appID }}
            show={updateGroupModalVisible}
            onHide={closeUpdateGroupModal}
          />
        )}
      </Paper>
    </>
  );
}

export default GroupList;
