import Grid from '@mui/material/Grid';

import { Channel } from '../../api/apiDataTypes';
import ChannelItem from '../Channels/ChannelItem';

function ApplicationItemChannelsList(props: { channels?: Channel[] }) {
  const channels = props.channels || [];

  return (
    <Grid container justifyContent="space-between">
      {channels.map(channel => (
        <Grid key={channel.id} size={4}>
          <ChannelItem channel={channel} />
        </Grid>
      ))}
    </Grid>
  );
}

export default ApplicationItemChannelsList;
