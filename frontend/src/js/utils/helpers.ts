import { Color, Theme } from '@material-ui/core';
import amber from '@material-ui/core/colors/amber';
import deepOrange from '@material-ui/core/colors/deepOrange';
import lime from '@material-ui/core/colors/lime';
import orange from '@material-ui/core/colors/orange';
import red from '@material-ui/core/colors/red';
import React from 'react';
import API from '../api/API';
import { Channel, Group } from '../api/apiDataTypes';

// Indexes/keys for the architectures need to match the ones in
// pkg/api/arches.go.
export const ARCHES: {
  [key: number]: string;
} = {
  0: 'ALL',
  1: 'AMD64',
  2: 'ARM64',
  3: 'X86',
};

export const ERROR_STATUS_CODE = 3;

const colors = makeColors();
export const timeIntervalsDefault = [
  { displayValue: '30 days', queryValue: '30d', disabled: false },
  { displayValue: '7 days', queryValue: '7d', disabled: false },
  { displayValue: '1 day', queryValue: '1d', disabled: false },
  { displayValue: '1 hour', queryValue: '1h', disabled: false },
];

export const defaultTimeInterval = timeIntervalsDefault.filter(
  interval => interval.queryValue === '1d'
)[0];

function makeColors() {
  const colors: string[] = [];

  const colorScheme = [lime, amber, orange, deepOrange, red];

  // Create a color list for versions to pick from.
  colorScheme.forEach((color: Color) => {
    // We choose the shades beyond 300 because they should not be too
    // light (in order to improve contrast).
    for (let i = 3; i <= 9; i += 2) {
      //@ts-ignore
      colors.push(color[i * 100]);
    }
  });

  return colors;
}

export function cleanSemverVersion(version: string) {
  let shortVersion = version;
  if (version.includes('+')) {
    shortVersion = version.split('+')[0];
  }
  return shortVersion;
}

export function getMinuteDifference(date1: number, date2: number) {
  return (date1 - date2) / 1000 / 60;
}

export function makeLocaleTime(
  timestamp: string | Date | number,
  opts: { useDate?: boolean; showTime?: boolean; dateFormat?: Intl.DateTimeFormatOptions } = {}
) {
  const { useDate = true, showTime = true, dateFormat = {} } = opts;
  const date = new Date(timestamp);
  const formattedDate = date.toLocaleDateString('default', dateFormat);
  const timeFormat = date.toLocaleString('default', { hour: '2-digit', minute: '2-digit' });

  if (useDate && showTime) {
    return `${formattedDate} ${timeFormat}`;
  }
  if (useDate) {
    return formattedDate;
  }
  return timeFormat;
}

export function makeColorsForVersions(
  theme: Theme,
  versions: string[],
  channel: Channel | null = null
) {
  const versionColors: { [key: string]: string } = {};
  let colorIndex = 0;
  let latestVersion = null;

  if (channel && channel.package) {
    latestVersion = cleanSemverVersion(channel.package.version);
  }

  for (let i = versions.length - 1; i >= 0; i--) {
    const version = versions[i];
    const cleanVersion = cleanSemverVersion(version);

    if (cleanVersion === latestVersion) {
      versionColors[cleanVersion] = theme.palette.primary.main;
    } else {
      versionColors[cleanVersion] = colors[colorIndex++ % colors.length];
    }
  }

  return versionColors;
}

export function getInstanceStatus(statusID: number, version?: string) {
  const status: {
    [x: number]: {
      type: string;
      className: string;
      spinning: boolean;
      icon: string;
      description: string;
      status: string;
      explanation: string;
      textColor?: string;
      bgColor?: string;
    };
  } = {
    1: {
      type: 'InstanceStatusUndefined',
      className: '',
      spinning: false,
      icon: '',
      description: '',
      status: 'Undefined',
      explanation: '',
    },
    2: {
      type: 'InstanceStatusUpdateGranted',
      className: 'warning',
      spinning: true,
      icon: '',
      description: 'Updating: granted',
      textColor: '#061751',
      bgColor: 'rgba(6, 23, 81, 0.1)',
      status: 'Granted',
      explanation:
        'The instance has received an update package -version ' +
        version +
        '- and the update process is about to start',
    },
    3: {
      type: 'InstanceStatusError',
      className: 'danger',
      spinning: false,
      icon: 'glyphicon glyphicon-remove',
      description: 'Error updating',
      bgColor: 'rgba(244, 67, 54, 0.1)',
      textColor: '#F44336',
      status: 'Error',
      explanation: 'The instance reported an error while updating to version ' + version,
    },
    4: {
      type: 'InstanceStatusComplete',
      className: 'success',
      spinning: false,
      icon: 'glyphicon glyphicon-ok',
      description: 'Update completed',
      status: 'Completed',
      textColor: '#26B640',
      bgColor: 'rgba(38, 182, 64, 0.1)',
      explanation:
        'The instance has been updated successfully and is now running version ' + version,
    },
    5: {
      type: 'InstanceStatusInstalled',
      className: 'warning',
      spinning: true,
      icon: '',
      description: 'Updating: installed',
      status: 'Installed',
      explanation:
        'The instance has installed the update package -version ' +
        version +
        '- but it isn’t using it yet',
    },
    6: {
      type: 'InstanceStatusDownloaded',
      className: 'warning',
      spinning: true,
      icon: '',
      description: 'Updating: downloaded',
      bgColor: 'rgba(44, 152, 240, 0.1)',
      textColor: '#2C98F0',
      status: 'Downloaded',
      explanation:
        'The instance has downloaded the update package -version ' +
        version +
        '- and will install it now',
    },
    7: {
      type: 'InstanceStatusDownloading',
      className: 'warning',
      spinning: true,
      icon: '',
      description: 'Updating: downloading',
      textColor: '#808080',
      bgColor: 'rgba(128, 128, 128, 0.1)',
      status: 'Downloading',
      explanation:
        'The instance has just started downloading the update package -version ' + version + '-',
    },
    8: {
      type: 'InstanceStatusOnHold',
      className: 'default',
      spinning: false,
      icon: '',
      description: 'Waiting...',
      status: 'On hold',
      explanation:
        'There was an update pending for the instance but it was put on hold because of the rollout policy',
    },
  };

  const statusDetails = statusID ? status[statusID] : status[1];

  return statusDetails;
}

export function useGroupVersionBreakdown(group: Group) {
  const [versionBreakdown, setVersionBreakdown] = React.useState([]);

  React.useEffect(() => {
    if (!group) {
      return;
    }

    API.getGroupVersionBreakdown(group.application_id, group.id)
      .then(version_breakdown => {
        setVersionBreakdown(version_breakdown || []);
      })
      .catch(err => {
        console.error('Error getting version breakdown for group', group.id, '\nError:', err);
      });
  }, [group]);

  return versionBreakdown;
}

// Keep in sync with https://github.com/flatcar-linux/update_engine/blob/flatcar-master/src/update_engine/action_processor.h#L25
const actionCodes: { [x: string]: string } = {
  '1': 'Error',
  '2': 'OmahaRequestError',
  '3': 'OmahaResponseHandlerError',
  '4': 'FilesystemCopierError',
  '5': 'PostinstallRunnerError',
  '6': 'SetBootableFlagError',
  '7': 'InstallDeviceOpenError',
  '8': 'KernelDeviceOpenError',
  '9': 'DownloadTransferError',
  '10': 'PayloadHashMismatchError',
  '11': 'PayloadSizeMismatchError',
  '12': 'DownloadPayloadVerificationError',
  '13': 'DownloadNewPartitionInfoError',
  '14': 'DownloadWriteError',
  '15': 'NewRootfsVerificationError',
  '16': 'NewKernelVerificationError',
  '17': 'SignedDeltaPayloadExpectedError',
  '18': 'DownloadPayloadPubKeyVerificationError',
  '19': 'PostinstallBootedFromFirmwareB',
  '20': 'DownloadStateInitializationError',
  '21': 'DownloadInvalidMetadataMagicString',
  '22': 'DownloadSignatureMissingInManifest',
  '23': 'DownloadManifestParseError',
  '24': 'DownloadMetadataSignatureError',
  '25': 'DownloadMetadataSignatureVerificationError',
  '26': 'DownloadMetadataSignatureMismatch',
  '27': 'DownloadOperationHashVerificationError',
  '28': 'DownloadOperationExecutionError',
  '29': 'DownloadOperationHashMismatch',
  '30': 'OmahaRequestEmptyResponseError',
  '31': 'OmahaRequestXMLParseError',
  '32': 'DownloadInvalidMetadataSize',
  '33': 'DownloadInvalidMetadataSignature',
  '34': 'OmahaResponseInvalid',
  '35': 'OmahaUpdateIgnoredPerPolicy',
  '36': 'OmahaUpdateDeferredPerPolicy',
  '37': 'OmahaErrorInHTTPResponse',
  '38': 'DownloadOperationHashMissingError',
  '39': 'DownloadMetadataSignatureMissingError',
  '40': 'OmahaUpdateDeferredForBackoff',
  '41': 'PostinstallPowerwashError',
  '42': 'NewPCRPolicyVerificationError',
  '43': 'NewPCRPolicyHTTPError',
  '44': 'RollbackError',
  '100': 'DownloadIncomplete',
  '2000': 'OmahaRequestHTTPResponseBase',
};

const flagsCodes: { [x: number]: string } = {
  [Math.pow(2, 31)]: 'DevModeFlag',
  [Math.pow(2, 30)]: 'ResumedFlag',
  [Math.pow(2, 29)]: 'TestImageFlag',
  [Math.pow(2, 28)]: 'TestOmahaUrlFlag',
};

export function getErrorAndFlags(errorCode: number) {
  const errorMessage = [];
  let errorCodeVal = errorCode;
  // Extract and remove flags from the error code
  var flags = [];
  for (const [flag, flagValue] of Object.entries(flagsCodes)) {
    if (errorCodeVal & parseInt(flag)) {
      errorCodeVal &= ~flag;
      flags.push(flagValue);
    }
  }
  if (actionCodes[errorCodeVal]) {
    errorMessage.push(actionCodes[errorCodeVal]);
  } else if (errorCodeVal > 2000 && errorCodeVal < 3000) {
    errorMessage.push(`Http error code(${errorCodeVal - 2000})`);
  } else {
    errorMessage.push(`Unknown Error ${errorCodeVal}`);
  }
  return [errorMessage, flags];
}

export function prepareErrorMessage(errorMessages: string[], flags: string[]) {
  const enhancedErrorMessages = errorMessages.reduce((acc, val) => `${val} ${acc}`, '');
  const enhancedFlags = flags.reduce((acc, val) => `${val} ${acc}`, '');
  return `${enhancedErrorMessages} ${enhancedFlags.length > 0 ? ' with ' + enhancedFlags : ''}`;
}
