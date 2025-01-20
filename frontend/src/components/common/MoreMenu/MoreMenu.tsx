import MoreVertIcon from '@mui/icons-material/MoreVert';
import IconButton from '@mui/material/IconButton';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import React from 'react';
import { useTranslation } from 'react-i18next';

let menuCount = 0;

interface MoreMenuProps {
  options: {
    label: string;
    action: () => void;
  }[];
  iconButtonProps?: React.ComponentProps<typeof IconButton>;
}

export default function MoreMenu(props: MoreMenuProps) {
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
        {...props.iconButtonProps}
        size="large"
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
