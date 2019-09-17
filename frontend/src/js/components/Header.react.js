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
import { Icon } from '@iconify/react';
import React from 'react';
import nebraskaLogo from '../icons/nebraska-logo.json';
import _ from 'underscore';

const useStyles = makeStyles(theme => ({
  title: {
    flexGrow: 1,
    display: 'none',
    [theme.breakpoints.up('sm')]: {
      display: 'block',
    },
  },
  header: {
    marginBottom: theme.spacing(1),
    background: process.env.APPBAR_BG || theme.palette.primary.main,
  },
}));

export default function Header() {
  const classes = useStyles();
  const projectLogo = _.isEmpty(nebraskaLogo) ? null : nebraskaLogo;

  let [menuAnchorEl, setMenuAnchorEl] = React.useState(null);

  function handleMenu(event) {
    setMenuAnchorEl(event.currentTarget);
  }

  function handleClose() {
    setMenuAnchorEl(null);
  }

  return (
    <AppBar position='static' className={classes.header}>
      <Toolbar>
        {projectLogo &&
          <Icon icon={projectLogo} height={45} />
        }
        <Typography variant='h6' className={classes.title}>
          {process.env.PROJECT_NAME}
        </Typography>
        <IconButton
          aria-label='User menu'
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
          <MenuItem
            component="a"
            href="https://github.com/settings/apps/authorizations"
          >
            <ListItemIcon>
              <CreateOutlined />
            </ListItemIcon>
            <ListItemText
              primary="Manage Access"/>
          </MenuItem>
        </Menu>
      </Toolbar>
    </AppBar>
  );
}
