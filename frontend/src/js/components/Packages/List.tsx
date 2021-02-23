import Box from '@material-ui/core/Box';
import MuiList from '@material-ui/core/List';
import Paper from '@material-ui/core/Paper';
import TablePagination from '@material-ui/core/TablePagination';
import PropTypes from 'prop-types';
import React from 'react';
import _ from 'underscore';
import API from '../../api/API';
import { Package } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import Empty from '../Common/EmptyContent';
import ListHeader from '../Common/ListHeader';
import Loader from '../Common/Loader';
import ModalButton from '../Common/ModalButton';
import EditDialog from './EditDialog';
import Item from './Item';

function List(props: {
  appID: string;
}) {
  const [application, setApplication] =
    React.useState(applicationsStore.getCachedApplication(props.appID) || null);
  const [packages, setPackages] = React.useState<Package[] | null>(null);
  const [packageToUpdate, setPackageToUpdate] = React.useState<Package | null>(null);
  const rowsPerPage = 10;
  const [page, setPage] = React.useState(0);

  function onChange() {
    setApplication(applicationsStore.getCachedApplication(props.appID));
  }

  React.useEffect(() => {
    applicationsStore.addChangeListener(onChange);
    if (!packages){
      API.getPackages(props.appID)
        .then((result) => {
          if (_.isNull(result)) {
            setPackages([]);
            return;
          }
          setPackages(result);
        });
    }

    if (application === null) {
      applicationsStore.getApplication(props.appID);
    }

    return function cleanup() {
      applicationsStore.removeChangeListener(onChange);
    };
  },
  [application, packages]);

  function onCloseEditDialog() {
    setPackageToUpdate(null);
  }

  function openEditDialog(packageID: string) {
    const pkg = packages?.find(({id}) => id === packageID) || null;
    if (pkg !== packageToUpdate) {
      setPackageToUpdate(pkg);
    }
  }

  function handleChangePage(event: React.MouseEvent<HTMLButtonElement, MouseEvent> | null,
                            newPage: number) {
    setPage(newPage);
  }

  return (
    <>
      <ListHeader
        title="Packages"
        actions={application ?
          [
            <ModalButton
              modalToOpen="AddPackageModal"
              data={{
                channels: application.channels || [],
                appID: props.appID
              }}
            />
          ]
          :
          []
        }
      />
      <Paper>

        <Box padding="1em">
          {application && !_.isNull(packages) ?
            _.isEmpty(packages) ?
              <Empty>This application does not have any package yet</Empty>
              :
              <React.Fragment>
                <MuiList>
                  {
                    packages.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
                      .map(packageItem =>
                        <Item
                          key={'packageItemID_' + packageItem.id}
                          packageItem={packageItem}
                          channels={application.channels}
                          handleUpdatePackage={openEditDialog}
                        />
                      )
                  }
                </MuiList>
                {packageToUpdate &&
                <EditDialog
                  data={{channels: application.channels, channel: packageToUpdate}}
                  show={Boolean(packageToUpdate)}
                  onHide={onCloseEditDialog}
                />
                }
                <TablePagination
                  rowsPerPageOptions={[]}
                  component="div"
                  count={packages.length}
                  rowsPerPage={rowsPerPage}
                  page={page}
                  backIconButtonProps={{
                    'aria-label': 'previous page',
                  }}
                  nextIconButtonProps={{
                    'aria-label': 'next page',
                  }}
                  onChangePage={handleChangePage}
                />
              </React.Fragment>
            :
              <Loader />
          }
        </Box>
      </Paper>
    </>
  );
}

List.propTypes = {
  appID: PropTypes.string.isRequired
};

export default List;
