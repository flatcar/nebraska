import React from 'react';
import Button from '@material-ui/core/Button';
import Icon from '@material-ui/core/Icon';
import Popover from '@material-ui/core/Popover';
import { TwitterPicker } from 'react-color';

// @todo: This needs to become a FormControl so we can display it in a similar
// style as the other form controls.

export function ColorPickerButton(props) {
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
      <Button variant="outlined" onClick={handleColorButtonClick}>
        <Icon style={{ backgroundColor: channelColor }} />
      </Button>
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