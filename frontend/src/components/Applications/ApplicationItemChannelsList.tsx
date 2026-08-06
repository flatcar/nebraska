import Grid from '@mui/material/Grid';
import React from 'react';

import { Channel } from '../../api/apiDataTypes';
import ChannelItem from '../Channels/ChannelItem';

function ApplicationItemChannelsList(props: { channels?: Channel[] }) {
  const channels = props.channels || [];
  let entries: React.ReactNode[] = [];

  if (channels) {
    entries = channels.map((channel, i) => (
      <ChannelItem channel={channel} key={`channelItem_${i}`} />
    ));
  }

  return (
    <Grid container justifyContent="space-between">
      {entries.map((entry: any, i: number) => (
        <Grid key={entry?.props?.channel?.id || entry?.key || i} size={4}>
          {entry}
        </Grid>
      ))}
    </Grid>
  );
}

export default ApplicationItemChannelsList;
