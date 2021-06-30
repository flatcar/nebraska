import alertCircleOutline from '@iconify/icons-mdi/alert-circle-outline';
import checkCircleOutline from '@iconify/icons-mdi/check-circle-outline';
import downloadCircleOutline from '@iconify/icons-mdi/download-circle-outline';
import helpCircleOutline from '@iconify/icons-mdi/help-circle-outline';
import packageVariantClosed from '@iconify/icons-mdi/package-variant-closed';
import pauseCircle from '@iconify/icons-mdi/pause-circle';
import playCircle from '@iconify/icons-mdi/play-circle';
import progressDownload from '@iconify/icons-mdi/progress-download';
import { Theme } from '@material-ui/core';

function makeStatusDefs(
  theme: Theme
): {
  [key: string]: {
    label: string;
    color: string;
    icon: any;
    queryValue: string;
  };
} {
  return {
    InstanceStatusComplete: {
      label: 'Complete',
      color: 'rgba(15,15,15,1)',
      icon: checkCircleOutline,
      queryValue: '4',
    },
    InstanceStatusDownloaded: {
      label: 'Downloaded',
      color: 'rgba(40,95,43,1)',
      icon: downloadCircleOutline,
      queryValue: '6',
    },
    InstanceStatusOnHold: {
      label: 'On Hold',
      color: theme.palette.grey['400'],
      icon: pauseCircle,
      queryValue: '8',
    },
    InstanceStatusInstalled: {
      label: 'Installed',
      color: 'rgba(27,92,145,1)',
      icon: packageVariantClosed,
      queryValue: '5',
    },
    InstanceStatusDownloading: {
      label: 'Downloading',
      color: 'rgba(17,40,141,1)',
      icon: progressDownload,
      queryValue: '7',
    },
    InstanceStatusError: {
      label: 'Error',
      color: 'rgba(164,45,36,1)',
      icon: alertCircleOutline,
      queryValue: '3',
    },
    InstanceStatusUndefined: {
      label: 'Unknown',
      color: 'rgb(89, 89, 89)',
      icon: helpCircleOutline,
      queryValue: '1',
    },
    InstanceStatusUpdateGranted: {
      label: 'Update Granted',
      color: theme.palette.sapphireColor,
      icon: playCircle,
      queryValue: '2',
    },
  };
}

export default makeStatusDefs;
