import AddIcon from '@mui/icons-material/Add';
import { IconButton } from '@mui/material';
import PropTypes from 'prop-types';
import React from 'react';
import { useTranslation } from 'react-i18next';

import ApplicationEdit from '../../Applications/ApplicationEdit';
import ChannelEdit from '../../Channels/ChannelEdit';
import GroupEditDialog from '../../Groups/GroupEditDialog/GroupEditDialog';
import PackageEditDialog, {
  EditDialogProps as PackageEditDialogProps,
} from '../../Packages/EditDialog';

interface ModalButtonProps {
  modalToOpen: string;
  data: { [key: string]: any };
}

function ModalButton(props: ModalButtonProps) {
  const [showModal, setShowModal] = React.useState(false);
  const { t } = useTranslation();

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
      modal = <ApplicationEdit {...options} />;
      break;
    case 'AddGroupModal':
      modal = <GroupEditDialog {...options} />;
      break;
    case 'AddChannelModal':
      modal = <ChannelEdit {...options} />;
      break;
    case 'AddPackageModal':
      modal = <PackageEditDialog {...(options as PackageEditDialogProps)} />;
      break;
  }

  // @todo: verify whether aria-label should be more specific (in which
  // case it should be set from the caller).
  return (
    <div>
      <IconButton
        size="small"
        aria-label={t('frequent|add_lower')}
        onClick={open}
        data-testid="modal-button"
      >
        <AddIcon fontSize="large" sx={{ color: theme => theme.palette.titleColor }} />
      </IconButton>
      {modal}
    </div>
  );
}

ModalButton.propTypes = {
  modalToOpen: PropTypes.string.isRequired,
  data: PropTypes.object,
};

export default ModalButton;
