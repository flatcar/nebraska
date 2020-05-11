import IconButton from '@material-ui/core/IconButton';
import Popover from '@material-ui/core/Popover';
import { makeStyles } from '@material-ui/styles';
import React from 'react';
import { TwitterPicker } from 'react-color';
import ChannelAvatar from '../Channels/ChannelAvatar';

// @todo: This needs to become a FormControl so we can display it in a similar
// style as the other form controls.

const useStyles = makeStyles({
  iconButton: {
    padding: '0',
  },
});

export function ColorPickerButton(props) {
  const classes = useStyles();
  const [channelColor, setChannelColor] = React.useState(props.color);
  const [displayColorPicker, setDisplayColorPicker] = React.useState(false);
  const [anchorEl, setAnchorEl] = React.useState(null);
  const {onColorPicked, componentColorProp = null} = props;

  const componentProps = {};
  componentProps[componentColorProp] = channelColor;

  function handleColorChange(color) {
    setChannelColor(color.hex);
    onColorPicked(color);
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
      <IconButton className={classes.iconButton}
        onClick={handleColorButtonClick}
        data-testid="icon-button"
      >
        {props.children ?
          React.cloneElement(props.children, componentProps)
          :
          <ChannelAvatar color={channelColor} />
        }
      </IconButton>
      {displayColorPicker &&
      <Popover
        data-testid="popover"
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
