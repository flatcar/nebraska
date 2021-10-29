import { Meta } from '@storybook/react/types-6-0';
import ChannelAvatar from './ChannelAvatar';

export default {
  title: 'channels/ChannelAvatar',
} as Meta;

export const White = () => <ChannelAvatar color="#fff">ABC</ChannelAvatar>;
