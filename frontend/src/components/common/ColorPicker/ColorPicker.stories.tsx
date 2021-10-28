import { Meta, Story } from '@storybook/react/types-6-0';
import ChannelAvatar from '../../Channels/ChannelAvatar';
import ColorPicker, { ColorPickerProps } from './ColorPicker';

export default {
  title: 'ColorPickerButton',
  argTypes: {
    onColorPicked: { action: 'onColorPicked' },
  },
} as Meta;

const ColorPickerButtonTemplate: Story<ColorPickerProps> = args => <ColorPicker {...args} />;
export const Closed = ColorPickerButtonTemplate.bind({});
Closed.args = {
  color: '#EB144C',
  children: <ChannelAvatar>Beta</ChannelAvatar>,
  initialOpen: false,
};

export const Open = ColorPickerButtonTemplate.bind({});
Open.args = {
  ...Closed.args,
  initialOpen: true,
};
