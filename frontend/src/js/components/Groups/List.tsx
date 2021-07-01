import { withStyles } from '@material-ui/core';
import MuiList from '@material-ui/core/List';
import Paper from '@material-ui/core/Paper';
import React from 'react';
import { Trans,useTranslation } from 'react-i18next';
import _ from 'underscore';
import { Channel, Group } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import Empty from '../Common/EmptyContent';
import ListHeader from '../Common/ListHeader';
import Loader from '../Common/Loader';
import ModalButton from '../Common/ModalButton';
import EditDialog from './EditDialog';
import Item from './Item';

const styles = () => ({
  root: {
    '& > hr:first-child': {
      display: 'none',
    },
  },
});

function List(props: { appID: string; classes: Record<'root', string> }) {
  const [application, setApplication] = React.useState(
    applicationsStore.getCachedApplication(props.appID)
  );
  const [searchTerm, setSearchTerm] = React.useState('');
  const [updateGroupModalVisible, setUpdateGroupModalVisible] = React.useState(false);
  const [updateGroupIDModal, setUpdateGroupIDModal] = React.useState<string | null>(null);
  const [updateAppIDModal, setUpdateAppIDModal] = React.useState<string | null>(null);
  const { t } = useTranslation();

  function closeUpdateGroupModal() {
    setUpdateGroupModalVisible(false);
  }

  function openUpdateGroupModal(appID: string, groupID: string) {
    setUpdateGroupModalVisible(true);
    setUpdateGroupIDModal(groupID);
    setUpdateAppIDModal(appID);
  }

  React.useEffect(() => {
    applicationsStore.addChangeListener(onChange);
    return () => {
      applicationsStore.removeChangeListener(onChange);
    };
  }, []);

  function onChange() {
    setApplication(applicationsStore.getCachedApplication(props.appID));
  }

  function searchUpdated(event: React.ChangeEvent<{ value: any }>) {
    const { value } = event.currentTarget;
    setSearchTerm(value.toLowerCase());
  }

  let channels: Channel[] = [];
  let groups: Group[] = [];
  let packages = [];
  let instances = 0;
  let name = '';
  let entries: React.ReactNode = '';

  if (application) {
    name = application.name;
    groups = application.groups ? application.groups : [];
    packages = application.packages ? application.packages : [];
    instances = application.instances ? application.instances : [];
    channels = application.channels ? application.channels : [];

    if (searchTerm) {
      groups = groups.filter((app: Group) => app.name.toLowerCase().includes(searchTerm));
    }

    if (_.isEmpty(groups)) {
      if (searchTerm) {
        entries = <Empty>{t('groups|No results found.')}</Empty>;
      } else {
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
      }
    } else {
      entries = _.map(groups, (group, i) => {
        return (
          <Item
            key={'groupID_' + group.id}
            group={group}
            appName={name}
            channels={channels}
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
  const { classes } = props;

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
              appID: props.appID,
            }}
          />,
        ]}
      />
      <Paper>
        <MuiList className={classes.root}>{entries}</MuiList>
        {groupToUpdate && (
          <EditDialog
            data={{ group: groupToUpdate, channels: channels }}
            show={updateGroupModalVisible}
            onHide={closeUpdateGroupModal}
          />
        )}
      </Paper>
    </>
  );
}

export default withStyles(styles)(List);
