import React from 'react';
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
      Are you sure ... ?
      <p className="button-group">
        <button type="button" className="confirm-dialog-btn-abord">
          No
        </button>
        <button type="button" className="confirm-dialog-btn-confirm" onClick={processClick}>
          Yes
        </button>
      </p>
    </div>
  );
}

export default ConfirmationContent;
