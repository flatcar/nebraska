import IconButton from '@material-ui/core/IconButton';
import Menu from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';
import MoreVertIcon from '@material-ui/icons/MoreVert';
import React from 'react';

export default function MoreMenu(props) {
  const [anchorEl, setAnchorEl] = React.useState(null);
  const options = props.options || [];

  function handleClick(event) {
    setAnchorEl(event.currentTarget);
  }

  function handleClose() {
    setAnchorEl(null);
  }

  return (
    <div>
      <IconButton edge="end"
        aria-controls="simple-menu"
        aria-haspopup="true"
        onClick={handleClick}
        data-testid="more-menu-open-button"
      >
        <MoreVertIcon />
      </IconButton>
      <Menu
        id="simple-menu"
        anchorEl={anchorEl}
        keepMounted
        open={Boolean(anchorEl)}
        onClose={handleClose}
      >
        {options.map(({label, action}, i) =>
          <MenuItem
            key={i}
            onClick={event => {
              handleClose(event);
              action();
            }}
            data-testid="more-menu-item"
          >
            {label}
          </MenuItem>
        )}
      </Menu>
    </div>
  );
}
