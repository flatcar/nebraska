import { List } from '@mui/material';
import Paper from '@mui/material/Paper';
import makeStyles from '@mui/styles/makeStyles';
import React from 'react';
import { Trans, useTranslation } from 'react-i18next';
import _ from 'underscore';
import { Application } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import Empty from '../common/EmptyContent';
import ListHeader from '../common/ListHeader';
import Loader from '../common/Loader';
import ModalButton from '../common/ModalButton';
import ApplicationEdit from './ApplicationEdit';
import ApplicationItem from './ApplicationItem';

const useStyles = makeStyles({
  root: {
    '& > hr:first-child': {
      display: 'none',
    },
  },
});

export interface ApplicationListProps {}

export default function ApplicationList() {
  const [applications, setApplications] = React.useState(
    applicationsStore().getCachedApplications ? applicationsStore().getCachedApplications() : []
  );

  React.useEffect(() => {
    applicationsStore().addChangeListener(onChange);
    return () => {
      applicationsStore().removeChangeListener(onChange);
    };
  }, []);

  React.useEffect(() => {
    if (applicationsStore().getCachedApplications) {
      setApplications(applicationsStore().getCachedApplications());
    }
  }, [applicationsStore().getCachedApplications]);

  function onChange() {
    setApplications(applicationsStore().getCachedApplications());
  }

  return <ApplicationListPure applications={applications} loading={applications === null} />;
}

export interface ApplicationListPureProps {
  /** To show. */
  applications: null | Application[];
  /** If we are waiting for applications to load. */
  loading?: boolean;
  /** If the edit screen is open for editId */
  editOpen?: boolean;
  /** The id to show for editing. */
  editId?: string;
  /** A default term to search. */
  defaultSearchTerm?: string;
}

export function ApplicationListPure(props: ApplicationListPureProps) {
  const classes = useStyles();
  const { t } = useTranslation();
  const [editOpen, setEditOpen] = React.useState(!!props.editOpen);
  const [editId, setEditId] = React.useState<null | string>(props.editId ? props.editId : null);
  const [searchTerm] = React.useState(props.defaultSearchTerm);

  function closeUpdateAppModal() {
    setEditOpen(false);
  }

  function openUpdateAppModal(appID: string) {
    setEditOpen(true);
    setEditId(appID);
  }

  let entries: React.ReactNode = '';
  const applications = props.applications
    ? searchTerm
      ? props.applications.filter(app => app.name.toLowerCase().includes(searchTerm))
      : props.applications
    : null;

  if (props.loading || applications === null) {
    entries = <Loader />;
  } else {
    if (_.isEmpty(applications)) {
      if (searchTerm) {
        entries = <Empty>{t('applications|No results found.')}</Empty>;
      } else {
        entries = (
          <Empty>
            <Trans ns="applications">
              Oops, it looks like you have not created any application yet..
              <br />
              <br /> Now is a great time to create your first one, just click on the plus symbol
              above.
            </Trans>
          </Empty>
        );
      }
    } else {
      entries = _.map(applications, (application: Application) => {
        return (
          <ApplicationItem
            description={application.description}
            groups={application.groups}
            id={application.id}
            key={application.id}
            name={application.name}
            numberOfInstances={application.instances?.count || 0}
            onUpdate={openUpdateAppModal}
            productId={application.product_id || ''}
          />
        );
      });
    }
  }

  const appToUpdate = applications && editId ? _.findWhere(applications, { id: editId }) : null;
  return (
    <>
      <ListHeader
        title={t('applications|Applications')}
        actions={[
          <ModalButton modalToOpen="AddApplicationModal" data={{ applications: applications }} />,
        ]}
      />
      <Paper>
        <List className={classes.root}>{entries}</List>
        {appToUpdate && (
          <ApplicationEdit data={appToUpdate} show={editOpen} onHide={closeUpdateAppModal} />
        )}
      </Paper>
    </>
  );
}
