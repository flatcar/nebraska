import { Icon, IconifyIcon } from '@iconify/react';

import alertCircleOutline from '../../icons/mdi/alert-circle-outline.json';
import alertOutline from '../../icons/mdi/alert-outline.json';
import checkCircleOutline from '../../icons/mdi/check-circle-outline.json';

export interface ActivityItemIconProps {
  severityName?: string;
}

export default function ActivityItemIcon(props: ActivityItemIconProps) {
  const { severityName } = props;
  const stateIcon = stateIcons[severityName || 'info'];

  return <Icon icon={stateIcon.icon} color={stateIcon.color} width="30px" height="30px" />;
}

const stateIcons: {
  [key: string]: {
    icon: IconifyIcon;
    color: string;
  };
} = {
  warning: {
    icon: alertOutline,
    color: '#ff5500',
  },
  info: {
    icon: alertCircleOutline,
    color: '#00d3ff',
  },
  error: {
    icon: alertCircleOutline,
    color: '#F44336',
  },
  success: {
    icon: checkCircleOutline,
    color: '#22bb00',
  },
};
