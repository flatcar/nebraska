import React from 'react';
import IconButton from '@material-ui/core/IconButton';
import MoreVertIcon from '@material-ui/icons/MoreVert';
import Menu from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';

export default function MoreMenu(props) {
  const [anchorEl, setAnchorEl] = React.useState(null);
  let options = props.options || [];

  function handleClick(event) {
    setAnchorEl(event.currentTarget);
  }

  function handleClose() {
    setAnchorEl(null);
  }

  return (
    <div>
      <IconButton edge="end" aria-controls="simple-menu" aria-haspopup="true" onClick={handleClick}>
        <MoreVertIcon />
      </IconButton>
      <Menu
        id="simple-menu"
        anchorEl={anchorEl}
        keepMounted
        open={Boolean(anchorEl)}
        onClose={handleClose}
      >
        {options.map(({label, action}) =>
          <MenuItem
            onClick={event => {
              handleClose(event);
              action();
          }}
          >
            {label}
          </MenuItem>
        )}
      </Menu>
    </div>
  );
}
