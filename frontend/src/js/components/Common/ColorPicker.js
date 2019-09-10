import React from 'react';
import IconButton from '@material-ui/core/IconButton';
import { makeStyles } from '@material-ui/styles';
import ChannelAvatar from '../Channels/ChannelAvatar';
import Popover from '@material-ui/core/Popover';
import { TwitterPicker } from 'react-color';

// @todo: This needs to become a FormControl so we can display it in a similar
// style as the other form controls.

const useStyles = makeStyles({
  iconButton: {
    padding: '0',
  },
});

export function ColorPickerButton(props) {
  const classes = useStyles();
  let [channelColor, setChannelColor] = React.useState(props.color);
  let [displayColorPicker, setDisplayColorPicker] = React.useState(false);
  let [anchorEl, setAnchorEl] = React.useState(null);
  let {color, onColorPicked} = props;

  function handleColorChange(color) {
    setChannelColor(color.hex);
    onColorPicked(color);
    color = color.hex;
  }

  function handleColorButtonClick(event) {
    setAnchorEl(event.currentTarget);
    setDisplayColorPicker(true);
  }

  function handleClose() {
    setAnchorEl(null);
    setDisplayColorPicker(false);
  }

  return (
    <div>
      <IconButton className={classes.iconButton} onClick={handleColorButtonClick}>
        <ChannelAvatar color={channelColor} />
      </IconButton>
      {displayColorPicker &&
      <Popover
        open={displayColorPicker}
        anchorEl={anchorEl}
        onClose={handleClose}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'center',
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'center',
        }}
      >
        <TwitterPicker color={channelColor} onChangeComplete={handleColorChange} triangle="hide"/>
      </Popover>
      }
    </div>
  );
}