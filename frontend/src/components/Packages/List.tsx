import Box from '@mui/material/Box';
import MuiList from '@mui/material/List';
import Paper from '@mui/material/Paper';
import TablePagination from '@mui/material/TablePagination';
import React from 'react';
import { useTranslation } from 'react-i18next';

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
  const [packageToUpdate, setPackageToUpdate] = React.useState<Package | null>(null);
  const { t } = useTranslation();
  const packageQueryParams = applicationsStore().getPackageQueryParams();

  function onChange() {
    setApplication(applicationsStore().getCachedApplication(props.appID));
  }

  React.useEffect(() => {
    applicationsStore().addChangeListener(onChange);

    if (application === null) {
      applicationsStore().getApplication(props.appID);
    }

    //eslint-disable-next-line eqeqeq
    if (application?.packages == null) {
      applicationsStore().getAndUpdatePackages(props.appID);
    }

    return function cleanup() {
      applicationsStore().removeChangeListener(onChange);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.appID, application]);

  function onCloseEditDialog() {
    setPackageToUpdate(null);
  }

  function handlePackageUpdated(updatedPackage: Package) {
    // Update the specific package in our local list
    setPackages(prevPackages => {
      if (!prevPackages) return prevPackages;
      return prevPackages.map(pkg => (pkg.id === updatedPackage.id ? updatedPackage : pkg));
    });
  }

  function openEditDialog(packageID: string) {
    const pkg = application?.packages?.items.find(({ id }) => id === packageID) || null;
    if (pkg !== packageToUpdate) {
      setPackageToUpdate(pkg);
    }
  }

  function handleChangePage(
    _event: React.MouseEvent<HTMLButtonElement, MouseEvent> | null,
    newPage: number
  ) {
    applicationsStore().setPackageQueryParams(
      { ...packageQueryParams, page: newPage },
      props.appID
    );
  }

  //eslint-disable-next-line eqeqeq
  const packagesLoading = application?.packages == null;

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
          {!packagesLoading ? (
            application?.packages?.totalCount === 0 ? (
              <Empty>This application does not have any package yet</Empty>
            ) : (
              <React.Fragment>
                <MuiList>
                  {application?.packages?.items.map(packageItem => (
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
                    onPackageUpdated={handlePackageUpdated}
                  />
                )}
                <TablePagination
                  rowsPerPageOptions={[]}
                  component="div"
                  count={application.packages?.totalCount || 0}
                  rowsPerPage={packageQueryParams.perPage}
                  page={packageQueryParams.page}
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
