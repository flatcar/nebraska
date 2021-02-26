import { List as MuiList, withStyles } from '@material-ui/core';
import Paper from '@material-ui/core/Paper';
import React from 'react';
import _ from 'underscore';
import { Application } from '../../api/apiDataTypes';
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

function List(props: { classes: Record<'root', string> }) {
  const [applications, setApplications] = React.useState(applicationsStore.getCachedApplications());
  const [searchTerm, setSearchTerm] = React.useState('');
  const [updateAppModalVisible, setUpdateModalVisible] = React.useState(false);
  const [updateAppIDModal, setUpdateAppIDModal] = React.useState<null | string>(null);

  function closeUpdateAppModal() {
    setUpdateModalVisible(false);
  }

  function openUpdateAppModal(appID: string) {
    setUpdateModalVisible(true);
    setUpdateAppIDModal(appID);
  }

  React.useEffect(() => {
    applicationsStore.addChangeListener(onChange);
    return () => {
      applicationsStore.removeChangeListener(onChange);
    };
  }, []);

  function onChange() {
    setApplications(applicationsStore.getCachedApplications());
  }

  let entries: React.ReactNode = '';
  const { classes } = props;

  if (searchTerm) {
    if (applications) {
      setApplications(applications.filter(app => app.name.toLowerCase().includes(searchTerm)));
    }
  }

  if (_.isNull(applications)) {
    entries = <Loader />;
  } else {
    if (_.isEmpty(applications)) {
      if (searchTerm) {
        entries = <Empty>No results found.</Empty>;
      } else {
        entries = (
          <Empty>
            Ops, it looks like you have not created any application yet..
            <br />
            <br /> Now is a great time to create your first one, just click on the plus symbol
            above.
          </Empty>
        );
      }
    } else {
      entries = _.map(applications, (application: Application, i: number) => {
        return (
          <Item
            key={application.id}
            application={application}
            handleUpdateApplication={openUpdateAppModal}
          />
        );
      });
    }
  }

  const appToUpdate =
    applications && updateAppIDModal ? _.findWhere(applications, { id: updateAppIDModal }) : null;
  return (
    <>
      <ListHeader
        title="Applications"
        actions={[
          <ModalButton modalToOpen="AddApplicationModal" data={{ applications: applications }} />,
        ]}
      />
      <Paper>
        <MuiList className={classes.root}>{entries}</MuiList>
        {appToUpdate && (
          <EditDialog
            data={appToUpdate}
            show={updateAppModalVisible}
            onHide={closeUpdateAppModal}
          />
        )}
      </Paper>
    </>
  );
}

export default withStyles(styles)(List);
