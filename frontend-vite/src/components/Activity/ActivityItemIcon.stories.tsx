import { Meta } from '@storybook/react';

import ActivityItemIcon from './ActivityItemIcon';

export default {
  title: 'activity/ActivityItemIcon',
} as Meta;

export const Default = () => <ActivityItemIcon />;
export const Warning = () => <ActivityItemIcon severityName="warning" />;
export const Info = () => <ActivityItemIcon severityName="info" />;
export const Error = () => <ActivityItemIcon severityName="error" />;
export const Success = () => <ActivityItemIcon severityName="success" />;
