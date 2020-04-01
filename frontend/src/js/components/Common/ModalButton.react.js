import Fab from '@material-ui/core/Fab';
import AddIcon from '@material-ui/icons/Add';
import PropTypes from 'prop-types';
import React from 'react';
import ApplicationEditDialog from '../Applications/EditDialog';
import ChannelEditDialog from '../Channels/EditDialog';
import GroupEditDialog from '../Groups/EditDialog';
import PackageEditDialog from '../Packages/EditDialog';

function ModalButton(props) {
  const [showModal, setShowModal] = React.useState(false);

  function close() {
    setShowModal(false);
  }

  function open() {
    setShowModal(true);
  }

  const options = {
    create: true,
    show: showModal,
    data: props.data,
    onHide: close,
  };

  let modal = null;
  switch (props.modalToOpen) {
    case 'AddApplicationModal':
      modal = <ApplicationEditDialog {...options} />;
      break;
    case 'AddGroupModal':
      modal = <GroupEditDialog {...options} />;
      break;
    case 'AddChannelModal':
      modal = <ChannelEditDialog {...options} />;
      break;
    case 'AddPackageModal':
      modal = <PackageEditDialog {...options} />;
      break;
  }

  // @todo: verify whether aria-label should be more specific (in which
  // case it should be set from the caller).
  return (
    <div>
      <Fab size="small" aria-label="add" onClick={open}>
        <AddIcon />
      </Fab>
      {modal}
    </div>
  );
}

ModalButton.propTypes = {
  modalToOpen: PropTypes.string.isRequired,
  data: PropTypes.object
};

export default ModalButton;
