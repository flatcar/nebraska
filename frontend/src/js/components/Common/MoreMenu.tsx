import IconButton from '@material-ui/core/IconButton';
import Menu from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';
import MoreVertIcon from '@material-ui/icons/MoreVert';
import React from 'react';
import { useTranslation } from 'react-i18next';

let menuCount = 0;

export default function MoreMenu(props: { options: { label: string; action: () => void }[] }) {
  const [anchorEl, setAnchorEl] = React.useState(null);
  const options = props.options || [];
  const { t } = useTranslation();
  const [menuId] = React.useState(() => {
    menuCount++;
    return `simple-menu-${menuCount}`;
  });

  function handleClick(event: any) {
    setAnchorEl(event.currentTarget);
  }

  function handleClose() {
    setAnchorEl(null);
  }

  return (
    <div>
      <IconButton
        edge="end"
        aria-controls={menuId}
        aria-haspopup="true"
        aria-label={t('common|Open menu')}
        onClick={handleClick}
        data-testid="more-menu-open-button"
      >
        <MoreVertIcon />
      </IconButton>
      <Menu
        id={menuId}
        anchorEl={anchorEl}
        keepMounted
        open={Boolean(anchorEl)}
        onClose={handleClose}
      >
        {options.map(({ label, action }, i) => (
          <MenuItem
            key={i}
            onClick={() => {
              handleClose();
              action();
            }}
            data-testid="more-menu-item"
          >
            {label}
          </MenuItem>
        ))}
      </Menu>
    </div>
  );
}
