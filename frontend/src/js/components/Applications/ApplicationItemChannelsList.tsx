import Grid from '@material-ui/core/Grid';
import React from 'react';
import { Channel } from '../../api/apiDataTypes';
import ChannelItem from '../Channels/Item';

function ApplicationItemChannelsList(props: {
  channels?: Channel[];

}) {
  const channels = props.channels || [];
  let entries: React.ReactNode[] = [];

  if (channels) {
    entries = channels.map((channel, i) =>
      <ChannelItem
        channel={channel}
        key={`channelItem_${i}`}
      />
    );
  }

  return (
    <Grid
      container
      justify="space-between"
    >
      {entries.map((entry: React.ReactNode, i: number) =>
        <Grid item xs={4} key={i}>
          {entry}
        </Grid>
      )}
    </Grid>
  );

}

export default ApplicationItemChannelsList;
