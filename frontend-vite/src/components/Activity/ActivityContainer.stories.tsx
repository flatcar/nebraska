import { Meta, StoryFn } from '@storybook/react';
import { MemoryRouter } from 'react-router-dom';

import { activityStore } from '../../stores/Stores';
import ActivityContainer, { ActivityContainerProps } from './ActivityContainer';

export default {
  title: 'activity/ActivityContainer',
} as Meta;

const TemplateEmpty: StoryFn<ActivityContainerProps> = args => {
  activityStore(true).stopRefreshing();
  activityStore(true).setActivity([]);
  return (
    <MemoryRouter>
      <ActivityContainer {...args} />
    </MemoryRouter>
  );
};

export const Empty = {
  render: TemplateEmpty,
  args: {},
};

const TemplateList: StoryFn<ActivityContainerProps> = args => {
  activityStore(true).stopRefreshing();
  activityStore(true).setActivity([
    {
      id: 1,
      app_id: '',
      group_id: '',
      created_ts: '2020-05-13T20:26:03.837688+05:30',
      class: 6,
      severity: 2,
      version: '0.0.0',
      application_name: 'ABC',
      group_name: null,
      channel_name: 'beta',
      instance_id: null,
    },
    {
      id: 2,
      app_id: '',
      group_id: '',
      created_ts: '2020-05-13T20:25:52.589886+05:30',
      class: 6,
      severity: 2,
      version: '0.0.0',
      application_name: 'DEF',
      group_name: null,
      channel_name: 'beta',
      instance_id: null,
    },
  ]);
  return (
    <MemoryRouter>
      <ActivityContainer {...args} />
    </MemoryRouter>
  );
};

export const List = {
  render: TemplateList,
  args: {},
};

const TemplateMultipleDays: StoryFn<ActivityContainerProps> = args => {
  activityStore(true).stopRefreshing();
  activityStore(true).setActivity([
    {
      id: 1,
      app_id: '',
      group_id: '',
      created_ts: '2020-05-13T20:26:03.837688+05:30',
      class: 6,
      severity: 2,
      version: '0.0.0',
      application_name: 'ABC',
      group_name: null,
      channel_name: 'beta',
      instance_id: null,
    },
    {
      id: 2,
      app_id: '',
      group_id: '',
      created_ts: '2020-05-13T20:25:52.589886+05:30',
      class: 6,
      severity: 2,
      version: '0.0.0',
      application_name: 'DEF',
      group_name: null,
      channel_name: 'beta',
      instance_id: null,
    },
    {
      id: 3,
      app_id: '',
      group_id: '',
      created_ts: '2020-05-14T20:26:03.837688+05:30',
      class: 6,
      severity: 1,
      version: '0.0.1',
      application_name: 'DEB',
      group_name: null,
      channel_name: 'beta',
      instance_id: null,
    },
    {
      id: 4,
      app_id: '',
      group_id: '',
      created_ts: '2020-05-14T20:25:52.589886+05:30',
      class: 6,
      severity: 3,
      version: '0.0.1',
      application_name: 'DOF',
      group_name: null,
      channel_name: 'beta',
      instance_id: null,
    },
    {
      id: 5,
      app_id: '',
      group_id: '',
      created_ts: '2020-05-14T20:26:03.837688+05:30',
      class: 6,
      severity: 1,
      version: '0.0.1',
      application_name: 'DAL',
      group_name: null,
      channel_name: 'alpha',
      instance_id: null,
    },
    {
      id: 6,
      app_id: '',
      group_id: '',
      created_ts: '2020-05-14T20:25:52.589886+05:30',
      class: 6,
      severity: 3,
      version: '0.0.1',
      application_name: 'DOP',
      group_name: null,
      channel_name: 'alpha',
      instance_id: null,
    },
  ]);
  return (
    <MemoryRouter>
      <ActivityContainer {...args} />
    </MemoryRouter>
  );
};

export const MultipleDays = {
  render: TemplateMultipleDays,
  args: {},
};
