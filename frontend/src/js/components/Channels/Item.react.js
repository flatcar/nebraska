import Avatar from '@material-ui/core/Avatar';
import ListItem from '@material-ui/core/ListItem';
import ListItemAvatar from '@material-ui/core/ListItemAvatar';
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction';
import ListItemText from '@material-ui/core/ListItemText';
import makeStyles from '@material-ui/styles/makeStyles';
import PropTypes from 'prop-types';
import React from "react";
import { cleanSemverVersion } from "../../constants/helpers";
import { applicationsStore } from "../../stores/Stores";
import MoreMenu from '../Common/MoreMenu';

const useStyles = makeStyles({
  colorAvatar: props => ({
    color: props.color,
    backgroundColor: props.backgroundColor,
    width: '30px',
    height: '30px'
  }),
});

function ChannelAvatar(props) {
  const classes = useStyles({
    color: props.color,
    backgroundColor: props.color,
  });

  return (
    <Avatar className={classes.colorAvatar} />
  );
}

function Item(props) {
  const channel = props.channel;
  const name = channel ? channel.name : '';
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
    <ListItem>
      <ListItemAvatar>
        <ChannelAvatar color={channel.color}/>
      </ListItemAvatar>
      <ListItemText
        primary={name}
        secondary={version ? cleanSemverVersion(version) : null}
      />
      <ListItemSecondaryAction>
        <MoreMenu options={[
          {label: 'Edit', action: updateChannel},
          {label: 'Delete', action: deleteChannel}
        ]} />
      </ListItemSecondaryAction>
    </ListItem>
  );
}

Item.propTypes = {
  channel: PropTypes.object.isRequired,
  packages: PropTypes.array.isRequired,
  handleUpdateChannel: PropTypes.func.isRequired
}

export default Item
