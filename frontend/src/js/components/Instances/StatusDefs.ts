import alertCircleOutline from '@iconify/icons-mdi/alert-circle-outline';
import checkCircleOutline from '@iconify/icons-mdi/check-circle-outline';
import clockAlertOutline from '@iconify/icons-mdi/clock-alert-outline';
import closeCircleOutline from '@iconify/icons-mdi/close-circle-outline';
import packageVariantClosed from '@iconify/icons-mdi/package-variant-closed';
import pauseCircle from '@iconify/icons-mdi/pause-circle';
import playCircle from '@iconify/icons-mdi/play-circle';
import refresh from '@iconify/icons-mdi/refresh';
import { Theme } from '@material-ui/core';
import { useTranslation } from 'react-i18next';

function makeStatusDefs(theme: Theme): {
  [key: string]: {
    label: string;
    color: string;
    icon: any;
    queryValue: string;
  };
} {
  const { t } = useTranslation();

  return {
    InstanceStatusComplete: {
      label: t('instances|Up to date'),
      color: 'rgba(15,15,15,1)',
      icon: checkCircleOutline,
      queryValue: '4',
    },
    InstanceStatusDownloaded: {
      label: t('instances|Downloaded'),
      color: 'rgba(40,95,43,1)',
      icon: refresh,
      queryValue: '6',
    },
    InstanceStatusOnHold: {
      label: t('instances|On Hold'),
      color: 'rgb(89, 89, 89)',
      icon: pauseCircle,
      queryValue: '8',
    },
    InstanceStatusInstalled: {
      label: t('instances|Installed'),
      color: 'rgba(27,92,145,1)',
      icon: packageVariantClosed,
      queryValue: '5',
    },
    InstanceStatusDownloading: {
      label: t('instances|Downloading'),
      color: 'rgba(17,40,141,1)',
      icon: refresh,
      queryValue: '7',
    },
    InstanceStatusError: {
      label: t('instances|Error'),
      color: 'rgba(164,45,36,1)',
      icon: closeCircleOutline,
      queryValue: '3',
    },
    InstanceStatusUndefined: {
      label: t('instances|Unknown'),
      color: 'rgb(89, 89, 89)',
      icon: alertCircleOutline,
      queryValue: '1',
    },
    InstanceStatusUpdateGranted: {
      label: t('instances|Update Granted'),
      color: theme.palette.sapphireColor,
      icon: playCircle,
      queryValue: '2',
    },
    InstanceStatusTimedOut: {
      label: t('instances|Timed out'),
      color: 'rgba(164,45,36,1)',
      icon: clockAlertOutline,
      queryValue: '8',
    },
    InstanceStatusOtherVersions: {
      label: t('instances|Not updating to this version'),
      color: 'rgb(89, 89, 89)',
      icon: alertCircleOutline,
      queryValue: '',
    },
  };
}

export default makeStatusDefs;
