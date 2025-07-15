import Box from '@mui/material/Box';
import MuiList from '@mui/material/List';
import Paper from '@mui/material/Paper';
import TablePagination from '@mui/material/TablePagination';
import React from 'react';
import { useTranslation } from 'react-i18next';
import _ from 'underscore';

import API from '../../api/API';
import { Package } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import Empty from '../common/EmptyContent';
import ListHeader from '../common/ListHeader';
import Loader from '../common/Loader';
import ModalButton from '../common/ModalButton';
import EditDialog from './EditDialog';
import Item from './Item';

interface ListProps {
  appID: string;
}

function List(props: ListProps) {
  const [application, setApplication] = React.useState(
    applicationsStore().getCachedApplication(props.appID) || null
  );
  const [packages, setPackages] = React.useState<Package[] | null>(null);
  const [packageToUpdate, setPackageToUpdate] = React.useState<Package | null>(null);
  const rowsPerPage = 10;
  const [page, setPage] = React.useState(0);
  const { t } = useTranslation();

  function onChange() {
    setApplication(applicationsStore().getCachedApplication(props.appID));
  }

  React.useEffect(() => {
    applicationsStore().addChangeListener(onChange);
    if (!packages) {
      // @todo: Request the pagination according to the page configuration in the table below.
      API.getPackages(props.appID, '', { perpage: 1000 })
        .then(result => {
          if (_.isNull(result.packages)) {
            setPackages([]);
            return;
          }
          setPackages(result.packages);
        })
        .catch(err => {
          console.error('Error getting the packages in the Packages/List: ', err);
        });
    }

    if (application === null) {
      applicationsStore().getApplication(props.appID);
    }

    return function cleanup() {
      applicationsStore().removeChangeListener(onChange);
    };
  }, [props.appID, application]);

  function onCloseEditDialog() {
    setPackageToUpdate(null);
  }

  function openEditDialog(packageID: string) {
    const pkg = packages?.find(({ id }) => id === packageID) || null;
    if (pkg !== packageToUpdate) {
      setPackageToUpdate(pkg);
    }
  }

  function handleChangePage(
    _event: React.MouseEvent<HTMLButtonElement, MouseEvent> | null,
    newPage: number
  ) {
    setPage(newPage);
  }

  return (
    <>
      <ListHeader
        title={t('packages|packages')}
        actions={
          application
            ? [
                <ModalButton
                  modalToOpen="AddPackageModal"
                  data={{
                    channels: application.channels || [],
                    appID: props.appID,
                  }}
                />,
              ]
            : []
        }
      />
      <Paper>
        <Box padding="1em">
          {application && !_.isNull(packages) ? (
            _.isEmpty(packages) ? (
              <Empty>This application does not have any package yet</Empty>
            ) : (
              <React.Fragment>
                <MuiList>
                  {packages
                    .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
                    .map(packageItem => (
                      <Item
                        key={'packageItemID_' + packageItem.id}
                        packageItem={packageItem}
                        channels={application.channels}
                        handleUpdatePackage={openEditDialog}
                      />
                    ))}
                </MuiList>
                {packageToUpdate && (
                  <EditDialog
                    data={{
                      appID: application.id,
                      channels: application.channels,
                      package: packageToUpdate,
                    }}
                    show={Boolean(packageToUpdate)}
                    onHide={onCloseEditDialog}
                  />
                )}
                <TablePagination
                  rowsPerPageOptions={[]}
                  component="div"
                  count={packages.length}
                  rowsPerPage={rowsPerPage}
                  page={page}
                  backIconButtonProps={{
                    'aria-label': t('frequent|previous_page'),
                  }}
                  nextIconButtonProps={{
                    'aria-label': t('frequent|next_page'),
                  }}
                  onPageChange={handleChangePage}
                />
              </React.Fragment>
            )
          ) : (
            <Loader />
          )}
        </Box>
      </Paper>
    </>
  );
}

export default List;
