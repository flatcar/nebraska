import ListItem from '@material-ui/core/ListItem';
import ListItemAvatar from '@material-ui/core/ListItemAvatar';
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction';
import ListItemText from '@material-ui/core/ListItemText';
import PropTypes from 'prop-types';
import React from "react";
import { cleanSemverVersion } from "../../constants/helpers";
import { applicationsStore } from "../../stores/Stores";
import MoreMenu from '../Common/MoreMenu';
import ChannelAvatar from './ChannelAvatar';

function Item(props) {
  let {channel, packages, handleUpdateChannel, ...others} = props;
  const name = channel.name;
  const version = channel.package ? cleanSemverVersion(channel.package.version) : 'No package';

  function deleteChannel() {
    let confirmationText = 'Are you sure you want to delete this channel?';
    if (confirm(confirmationText)) {
      applicationsStore.deleteChannel(channel.application_id, channel.id);
    }
  }

  function updateChannel() {
    props.handleUpdateChannel(channel.id);
  }

  return (
    <ListItem {...others}>
      <ListItemAvatar>
        <ChannelAvatar color={channel.color}>{name[0]}</ChannelAvatar>
      </ListItemAvatar>
      <ListItemText
        primary={name}
        secondary={version ? cleanSemverVersion(version) : null}
      />
      {props.handleUpdateChannel &&
        <ListItemSecondaryAction>
          <MoreMenu options={[
            {label: 'Edit', action: updateChannel},
            {label: 'Delete', action: deleteChannel}
          ]} />
        </ListItemSecondaryAction>
      }
    </ListItem>
  );
}

Item.propTypes = {
  channel: PropTypes.object.isRequired,
}

export default Item
