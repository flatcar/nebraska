import CheckCircleOutlined from '@mui/icons-material/CheckCircleOutlined';
import CloudDownloadOutlined from '@mui/icons-material/CloudDownloadOutlined';
import DownloadOutlined from '@mui/icons-material/DownloadOutlined';
import ErrorOutlined from '@mui/icons-material/ErrorOutlined';
import HelpOutlined from '@mui/icons-material/HelpOutlined';
import Inventory2Outlined from '@mui/icons-material/Inventory2Outlined';
import PauseCircle from '@mui/icons-material/PauseCircle';
import PlayCircle from '@mui/icons-material/PlayCircle';
import { Theme } from '@mui/material';
import type React from 'react';
import { useTranslation } from 'react-i18next';

function makeStatusDefs(theme: Theme): {
  [key: string]: {
    label: string;
    color: string;
    icon: React.ElementType;
    queryValue: string;
  };
} {
  // eslint-disable-next-line react-hooks/rules-of-hooks
  const { t } = useTranslation();

  return {
    InstanceStatusComplete: {
      label: t('instances|complete'),
      color: 'rgba(15,15,15,1)',
      icon: CheckCircleOutlined,
      queryValue: '4',
    },
    InstanceStatusDownloaded: {
      label: t('instances|downloaded'),
      color: 'rgba(40,95,43,1)',
      icon: DownloadOutlined,
      queryValue: '6',
    },
    InstanceStatusOnHold: {
      label: t('instances|on_hold'),
      color: theme.palette.grey['400'],
      icon: PauseCircle,
      queryValue: '8',
    },
    InstanceStatusInstalled: {
      label: t('instances|installed'),
      color: 'rgba(27,92,145,1)',
      icon: Inventory2Outlined,
      queryValue: '5',
    },
    InstanceStatusDownloading: {
      label: t('instances|downloading'),
      color: 'rgba(17,40,141,1)',
      icon: CloudDownloadOutlined,
      queryValue: '7',
    },
    InstanceStatusError: {
      label: t('instances|error'),
      color: 'rgba(164,45,36,1)',
      icon: ErrorOutlined,
      queryValue: '3',
    },
    InstanceStatusUndefined: {
      label: t('instances|unknown'),
      color: 'rgb(89, 89, 89)',
      icon: HelpOutlined,
      queryValue: '1',
    },
    InstanceStatusUpdateGranted: {
      label: t('instances|update_granted'),
      color: theme.palette.sapphireColor,
      icon: PlayCircle,
      queryValue: '2',
    },
  };
}

export default makeStatusDefs;
