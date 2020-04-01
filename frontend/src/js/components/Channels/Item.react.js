import ListItem from '@material-ui/core/ListItem';
import ListItemAvatar from '@material-ui/core/ListItemAvatar';
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction';
import ListItemText from '@material-ui/core/ListItemText';
import PropTypes from 'prop-types';
import React from 'react';
import { ARCHES, cleanSemverVersion } from '../../constants/helpers';
import { applicationsStore } from '../../stores/Stores';
import MoreMenu from '../Common/MoreMenu';
import ChannelAvatar from './ChannelAvatar';

function Item(props) {
  const {channel, packages, handleUpdateChannel, showArch=true, ...others} = props;
  const name = channel.name;
  const version = channel.package ? cleanSemverVersion(channel.package.version) : 'No package';

  function deleteChannel() {
    const confirmationText = 'Are you sure you want to delete this channel?';
    if (window.confirm(confirmationText)) {
      applicationsStore.deleteChannel(channel.application_id, channel.id);
    }
  }

  function updateChannel() {
    props.handleUpdateChannel(channel.id);
  }

  function getSecondaryText() {
    let text = '';

    if (version) {
      text = cleanSemverVersion(version);
    }

    if (showArch) {
      if (text !== '') {
        text += ' ';
      }

      text += `(${ARCHES[channel.arch]})`;
    }

    return text;
  }

  return (
    <ListItem component="div" {...others}>
      <ListItemAvatar>
        <ChannelAvatar color={channel.color}>{name[0]}</ChannelAvatar>
      </ListItemAvatar>
      <ListItemText
        primary={name}

        secondary={getSecondaryText()}
      />
      {props.handleUpdateChannel &&
        <ListItemSecondaryAction>
          <MoreMenu options={[
            {label: 'Edit', action: updateChannel},
            {label: 'Delete', action: deleteChannel}
          ]}
          />
        </ListItemSecondaryAction>
      }
    </ListItem>
  );
}

Item.propTypes = {
  channel: PropTypes.object.isRequired,
};

export default Item;
