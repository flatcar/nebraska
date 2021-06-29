import React from 'react';
import { useTranslation } from 'react-i18next';
import { applicationsStore } from '../../stores/Stores';

function ConfirmationContent(props: {
  data: {
    appID: string;
    groupID?: string;
    channelID?: string;
    packageID?: string;
    type: string;
  };
}) {
  const { t } = useTranslation();

  function processClick() {
    if (props.data.type === 'application') {
      applicationsStore.deleteApplication(props.data.appID);
    } else if (props.data.type === 'group') {
      applicationsStore.deleteGroup(props.data.appID, props.data.groupID as string);
    } else if (props.data.type === 'channel') {
      applicationsStore.deleteChannel(props.data.appID, props.data.channelID as string);
    } else if (props.data.type === 'package') {
      applicationsStore.deletePackage(props.data.appID, props.data.packageID as string);
    }
  }

  return (
    <div className="popover-content" {...props}>
      {t('common|Are you sure ... ?')}
      <p className="button-group">
        <button type="button" className="confirm-dialog-btn-abord">
          {t('frequent|No')}
        </button>
        <button type="button" className="confirm-dialog-btn-confirm" onClick={processClick}>
          {t('frequent|Yes')}
        </button>
      </p>
    </div>
  );
}

export default ConfirmationContent;
