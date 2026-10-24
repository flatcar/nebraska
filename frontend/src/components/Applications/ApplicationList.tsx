import { List, TablePagination } from '@mui/material';
import Paper from '@mui/material/Paper';
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

export type ApplicationListProps = unknown;

export default function ApplicationList() {
  const [applications, setApplications] = React.useState(
    applicationsStore().getCachedApplications ? applicationsStore().getCachedApplications() : []
  );

  const [totalCount, setTotalCount] = React.useState(
    applicationsStore().getApplicationsTotalCount() || 0
  );

  const onChange = React.useCallback(() => {
    setApplications(applicationsStore().getCachedApplications());
    setTotalCount(applicationsStore().getApplicationsTotalCount());
  }, []);

  const applicationsQueryParams = applicationsStore().getApplicationsQueryParams();

  function handleChangePage(
    _event: React.MouseEvent<HTMLButtonElement, MouseEvent> | null,
    newPage: number
  ) {
    applicationsStore().setApplicationsQueryParams({ ...applicationsQueryParams, page: newPage });
  }

  React.useEffect(() => {
    // Get initial data in case store already has applications loaded
    const currentApplications = applicationsStore().getCachedApplications();
    if (currentApplications) {
      setApplications(currentApplications);
      setTotalCount(applicationsStore().getApplicationsTotalCount());
    }

    // Set up listener for future changes
    applicationsStore().addChangeListener(onChange);
    // Trigger initial fetch since stores don't fetch automatically
    applicationsStore().getApplications();
    return () => {
      applicationsStore().removeChangeListener(onChange);
    };
  }, [onChange]);

  return (
    <ApplicationListPure
      applicationsQueryParams={applicationsQueryParams}
      handleChangePage={handleChangePage}
      applications={applications}
      loading={applications === null}
      applicationsTotalCount={totalCount}
    />
  );
}

export interface ApplicationListPureProps {
  /** To show. */
  applications: null | Application[];

  applicationsTotalCount?: number;

  applicationsQueryParams?: { perPage: number; page: number };

  handleChangePage?: (event: React.MouseEvent<HTMLButtonElement> | null, newPage: number) => void;
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
        entries = <Empty>{t('applications|no_results_found')}</Empty>;
      } else {
        entries = (
          <Empty>
            <Trans t={t} ns="applications" i18nKey="noappyet">
              Oops, it looks like you have not created any application yet..
              <br />
              <br />
              Now is a great time to create your first one, just click on the plus symbol above.
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
        title={t('applications|applications')}
        actions={[
          <ModalButton modalToOpen="AddApplicationModal" data={{ applications: applications }} />,
        ]}
      />
      <Paper>
        <React.Fragment>
          <List
            sx={{
              '& > hr:first-of-type': {
                display: 'none',
              },
            }}
          >
            {entries}
          </List>
          {appToUpdate && (
            <ApplicationEdit data={appToUpdate} show={editOpen} onHide={closeUpdateAppModal} />
          )}
          {props.handleChangePage && props.applicationsQueryParams && (
            <TablePagination
              rowsPerPageOptions={[]}
              component="div"
              count={props.applicationsTotalCount ?? props.applications?.length ?? 0}
              rowsPerPage={props.applicationsQueryParams.perPage}
              page={props.applicationsQueryParams.page}
              backIconButtonProps={{
                'aria-label': t('frequent|previous_page'),
              }}
              nextIconButtonProps={{
                'aria-label': t('frequent|next_page'),
              }}
              onPageChange={props.handleChangePage}
            />
          )}
        </React.Fragment>
      </Paper>
    </>
  );
}
