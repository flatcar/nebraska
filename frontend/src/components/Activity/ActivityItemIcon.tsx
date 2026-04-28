import CheckCircleOutlined from '@mui/icons-material/CheckCircleOutlined';
import ErrorOutlined from '@mui/icons-material/ErrorOutlined';
import WarningAmberOutlined from '@mui/icons-material/WarningAmberOutlined';
import type React from 'react';

export interface ActivityItemIconProps {
  severityName?: string;
}

export default function ActivityItemIcon(props: ActivityItemIconProps) {
  const { severityName } = props;
  const stateIcon = stateIcons[severityName || 'info'];
  const IconComponent = stateIcon.icon;

  return <IconComponent sx={{ color: stateIcon.color, fontSize: 30 }} />;
}

const stateIcons: {
  [key: string]: {
    icon: React.ElementType;
    color: string;
  };
} = {
  warning: {
    icon: WarningAmberOutlined,
    color: '#ff5500',
  },
  info: {
    icon: ErrorOutlined,
    color: '#00d3ff',
  },
  error: {
    icon: ErrorOutlined,
    color: '#F44336',
  },
  success: {
    icon: CheckCircleOutlined,
    color: '#22bb00',
  },
};
