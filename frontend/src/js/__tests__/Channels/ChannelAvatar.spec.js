import {render} from '@testing-library/react';
import React from 'react';
import ChannelAvatar from '../../components/Channels/ChannelAvatar';

describe('Channel Avatar', () => {
  it('should render avatar with correct color', () => {
    const {asFragment} = render(<ChannelAvatar color="#fff">ABC</ChannelAvatar>);
    expect(asFragment()).toMatchSnapshot();
  });
});
