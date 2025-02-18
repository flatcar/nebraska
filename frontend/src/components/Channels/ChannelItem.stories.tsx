import { Meta, StoryFn } from '@storybook/react';
import ChannelItem, { ChannelItemProps } from './ChannelItem';

export default {
  title: 'channels/ChannelItem',
  argTypes: {
    onChannelUpdate: { action: 'onChannelUpdate' },
  },
} as Meta;

const Template: StoryFn<ChannelItemProps> = args => <ChannelItem {...args} />;

export const ShowArch = {
  render: Template,

  args: {
    channel: {
      id: '30b6ffa6-e6dc-4a01-bea6-9ce7f1a5bb34',
      name: 'edge',
      color: '#f4ab3b',
      created_ts: '2015-09-19T07:09:34.265754+02:00',
      application_id: 'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
      package_id: 'X-A-Y',
      package: {
        id: 'PACK_ID',
        type: 4,
        version: '1.11.3',
        url: 'https://github.com/kinvolk',
        filename: '',
        description: '',
        size: '',
        hash: '',
        created_ts: '2019-07-18T20:10:39.163326+05:30',
        channels_blacklist: null,
        application_id: 'df1c8bbb-f525-49ee-8c94-3ca548b42059',
        flatcar_action: null,
        arch: 0,
        extra_files: [],
      },
      arch: 2,
    },
    showArch: true,
  },
};

export const NoPackage = {
  render: Template,

  args: {
    channel: {
      id: '30b6ffa6-e6dc-4a01-bea6-9ce7f1a5bb34',
      name: 'edge',
      color: '#f4ab3b',
      created_ts: '2015-09-19T07:09:34.265754+02:00',
      application_id: 'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
      package_id: null,
      package: null,
      arch: 2,
    },
    showArch: false,
  },
};

export const NoArch = {
  render: Template,

  args: {
    ...ShowArch.args,
    showArch: false,
  },
};

export const AppView = {
  render: Template,

  args: {
    ...ShowArch.args,
    isAppView: true,
  },
};
