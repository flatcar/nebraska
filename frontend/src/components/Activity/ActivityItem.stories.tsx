import { Meta, StoryFn } from '@storybook/react';
import { MemoryRouter } from 'react-router-dom';

import { ActivityItemPure, ActivityItemPureProps } from './ActivityItem';

export default {
  title: 'activity/ActivityItem',
} as Meta;

const Template: StoryFn<ActivityItemPureProps> = args => (
  <MemoryRouter>
    <ActivityItemPure {...args} />
  </MemoryRouter>
);

export const Warning = {
  render: Template,

  args: {
    createdTs: '2020-05-13T20:26:03.837688+05:30',
    appId: 'XXXX-XXX',
    groupId: 'YYYY-YYYY',
    classType: 'someType',
    groupName: 'some group',
    appName: 'some app',
    description: 'A description',
    severityName: 'warning',
  },
};

export const ActivityChannelPackageUpdated = {
  render: Template,

  args: {
    createdTs: '2020-05-13T20:26:03.837688+05:30',
    appId: 'XXXX-XXX',
    groupId: 'YYYY-YYYY',
    classType: 'activityChannelPackageUpdated',
    groupName: 'some group',
    appName: 'some app',
    description: 'A description',
    severityName: 'info',
  },
};
