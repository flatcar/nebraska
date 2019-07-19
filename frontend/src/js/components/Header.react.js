import AppBar from '@material-ui/core/AppBar';
import IconButton from '@material-ui/core/IconButton';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import Menu from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';
import { makeStyles } from '@material-ui/core/styles';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import AccountCircle from '@material-ui/icons/AccountCircle';
import CreateOutlined from '@material-ui/icons/CreateOutlined';
import DirectionsRunOutlined from '@material-ui/icons/DirectionsRunOutlined';
import React from 'react';
import API from '../api/API';
import ModalUpdatePassword from './Common/ModalUpdatePassword.react';

const useStyles = makeStyles(theme => ({
  title: {
    flexGrow: 1,
    display: 'none',
    [theme.breakpoints.up('sm')]: {
      display: 'block',
    },
  }
}));

export default function Header() {
  const classes = useStyles();

  let [menuAnchorEl, setMenuAnchorEl] = React.useState(null);
  let [showModal, setShowModal] = React.useState(false);

  var options = {
    show: showModal
  }

  function logout() {
    API.logout();
  }

  function handleMenu(event) {
    setMenuAnchorEl(event.currentTarget);
  }

  function handleClose() {
    setMenuAnchorEl(null);
    setShowModal(false);
  }

  function handleChangePassword() {
    setMenuAnchorEl(null);
    setShowModal(true);
  }

  return (
    <AppBar position='static'>
      <Toolbar>
        <Typography variant='h6' className={classes.title}>
          Nebraska
        </Typography>
        <IconButton
          aria-label='Account of current user'
          aria-controls='menu-appbar'
          aria-haspopup='true'
          onClick={handleMenu}
          color='inherit'
        >
          <AccountCircle />
        </IconButton>
        <Menu
          id='customized-menu'
          anchorEl={menuAnchorEl}
          keepMounted
          open={Boolean(menuAnchorEl)}
          onClose={handleClose}
          anchorOrigin={{
            vertical: 'top',
            horizontal: 'right',
          }}
          transformOrigin={{
            vertical: 'top',
            horizontal: 'right',
          }}
        >
          <MenuItem onClick={handleChangePassword}>
            <ListItemIcon>
              <CreateOutlined />
            </ListItemIcon>
            <ListItemText primary='Change Password' />
            <ModalUpdatePassword {...options} onHide={handleClose} />
          </MenuItem>
          <MenuItem onClick={logout}>
            <ListItemIcon>
              <DirectionsRunOutlined />
            </ListItemIcon>
            <ListItemText primary='Log out' />
          </MenuItem>
        </Menu>
      </Toolbar>
    </AppBar>
  );
}
