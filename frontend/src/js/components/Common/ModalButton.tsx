import { IconButton, makeStyles } from '@material-ui/core';
import AddIcon from '@material-ui/icons/Add';
import PropTypes from 'prop-types';
import React from 'react';
import { useTranslation } from 'react-i18next';
import ApplicationEditDialog from '../Applications/EditDialog';
import ChannelEditDialog from '../Channels/EditDialog';
import GroupEditDialog from '../Groups/EditDialog';
import PackageEditDialog, {
  EditDialogProps as PackageEditDialogProps,
} from '../Packages/EditDialog';

const useStyles = makeStyles(theme => ({
  root: {
    color: theme.palette.titleColor,
  },
}));

function ModalButton(props: { data: object; modalToOpen: string; icon?: string }) {
  const [showModal, setShowModal] = React.useState(false);
  const classes = useStyles();
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
      modal = <ApplicationEditDialog {...options} />;
      break;
    case 'AddGroupModal':
      modal = <GroupEditDialog {...options} />;
      break;
    case 'AddChannelModal':
      modal = <ChannelEditDialog {...options} />;
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
        aria-label={t('frequent|add')}
        onClick={open}
        data-testid="modal-button"
      >
        <AddIcon fontSize="large" className={classes.root} />
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
