import { Meta, StoryObj } from '@storybook/react';

import ChannelAvatar from '../../Channels/ChannelAvatar';
import ColorPicker from './ColorPicker';

const meta: Meta<typeof ColorPicker> = {
  title: 'ColorPickerButton',
  component: ColorPicker,
  argTypes: {
    onColorPicked: { action: 'onColorPicked' },
  },
};
export default meta;

type Story = StoryObj<typeof ColorPicker>;

export const Closed: Story = {
  args: {
    color: '#EB144C',
    children: <ChannelAvatar>Beta</ChannelAvatar>,
    initialOpen: false,
  },
};

export const Open: Story = {
  args: {
    color: '#EB144C',
    children: <ChannelAvatar>Beta</ChannelAvatar>,
    initialOpen: true,
  },
};
