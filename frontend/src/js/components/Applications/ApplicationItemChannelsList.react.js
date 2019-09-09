import Grid from '@material-ui/core/Grid';
import PropTypes from 'prop-types';
import React from "react";
import ChannelItem from '../Channels/Item.react';

function ApplicationItemChannelsList(props) {
  let channels = props.channels || [];
  let entries = '-';

  if (channels) {
    entries = channels.map((channel, i) =>
      <ChannelItem
        channel={channel}
        ContainerComponent="span"
      />
    );
  }

  return(
    <Grid
      container
      justify="space-between"
    >
      {entries.map(entry =>
        <Grid item xs={4}>
          {entry}
        </Grid>
      )}
    </Grid>
  );

}

ApplicationItemChannelsList.propTypes = {
  channels: PropTypes.array.isRequired
}

export default ApplicationItemChannelsList
