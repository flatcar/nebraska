import { Icon } from '@iconify/react';
import { Box, Button } from '@material-ui/core';
import AppBar from '@material-ui/core/AppBar';
import IconButton from '@material-ui/core/IconButton';
import Menu, { MenuProps } from '@material-ui/core/Menu';
import { createMuiTheme, makeStyles, Theme, ThemeProvider, useTheme } from '@material-ui/core/styles';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import AccountCircle from '@material-ui/icons/AccountCircle';
import CreateOutlined from '@material-ui/icons/CreateOutlined';
import DOMPurify from 'dompurify';
import React from 'react';
import _ from 'underscore';
import nebraskaLogo from '../icons/nebraska-logo.json';
import { UserState } from '../stores/redux/actions';
import { useTypedSelector } from '../stores/redux/reducers';

const useStyles = makeStyles(theme => ({
  title: {
    flexGrow: 1,
    display: 'none',
    color: theme.palette.titleColor,
    [theme.breakpoints.up('sm')]: {
      display: 'block',
    },
  },
  header: {
    marginBottom: theme.spacing(1),
    backgroundColor:
      theme.palette.type === 'dark' ? theme.palette.common.black : theme.palette.common.white,
  },
  svgContainer: {
    '& svg': { maxHeight: '3rem' },
  },
  userName: {
    fontSize: '1.5em',
  },
  email: {
    fontSize: '1.2em',
  }
}));

function prepareDarkTheme(theme: Theme) {
  return createMuiTheme({
    ...theme,
    palette: {
      type: 'dark',
      primary: {
        contrastText: '#fff',
        main: '#000',
      },
    },
  });
}

interface NebraskaConfig {
  title?: string;
  logo?: string;
  access_management_url?: string;
  header_style?: string;
  [other: string]: any;
}

interface AppbarProps {
  config: NebraskaConfig | null;
  user: UserState | null;
  menuAnchorEl: MenuProps['anchorEl'];
  projectLogo: object;
  handleClose: MenuProps['onClose'];
  handleMenu: React.MouseEventHandler<HTMLButtonElement>;
}

function Appbar(props: AppbarProps) {
  const { config, user, menuAnchorEl, projectLogo, handleClose, handleMenu } = props;
  const classes = useStyles();

  React.useEffect(() => {
    document.title = (config?.title) || 'Nebraska';
  }, [config]);

  function showAccountMenu() {
    return config?.access_management_url || user?.name || user?.email;
  }

  const showAccountButton = showAccountMenu();

  return (
    <AppBar position="static" className={classes.header}>
      <Toolbar>
        {config?.logo ? (
          <Box className={classes.svgContainer}>
            <div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(config.logo) }} />
          </Box>
        ) : (
          <Icon icon={projectLogo} height={45} />
        )}
        {config?.title && (
          <Typography variant="h6" className={classes.title}>
            {config.title}
          </Typography>
        )}
        <div style={{ flex: '1 0 0' }} />
        {showAccountButton && (
          <IconButton
            aria-label="User menu"
            aria-controls="menu-appbar"
            aria-haspopup="true"
            onClick={handleMenu}
          >
            <AccountCircle />
          </IconButton>
        )}
        {showAccountButton &&
          <Menu
            id="customized-menu"
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
            <Box
              paddingY={2}
              paddingX={2}
              textAlign="center"
            >
              {user?.name && <Typography className={classes.userName}>{user.name}</Typography>}
              {user?.email && <Typography className={classes.email}>{user.email}</Typography>}
            </Box>
            <Box
              paddingY={1}
              paddingX={2}
              textAlign="center"
            >
              <Button
                component="a"
                startIcon={<CreateOutlined />}
                disabled={!config?.access_management_url}
                href={config?.access_management_url || ''}
              >
                Manage Account
              </Button>
            </Box>
          </Menu>
        }
      </Toolbar>
    </AppBar>
  );
}

export default function Header() {
  const {config, user} = useTypedSelector(state => ({config: state.config, user: state.user}));
  const theme = useTheme();
  const projectLogo = _.isEmpty(nebraskaLogo) ? null : nebraskaLogo;
  const [menuAnchorEl, setMenuAnchorEl] = React.useState<HTMLButtonElement | null>(null);

  function handleMenu(event: React.MouseEvent<HTMLButtonElement>) {
    setMenuAnchorEl(event.currentTarget);
  }

  function handleClose() {
    setMenuAnchorEl(null);
  }

  const props = {config, user, menuAnchorEl, projectLogo: projectLogo as object, handleClose, handleMenu} as AppbarProps;
  const appBar = <Appbar {...props} />
  // cachedConfig.appBarColor is for backward compatibility (the name used for the setting before).
  return config && (config.header_style === 'dark' || config.header_style === undefined && config.appBarColor === 'dark') ? (
    <ThemeProvider theme={prepareDarkTheme(theme)}>{appBar}</ThemeProvider>
  ) : (
    appBar
  );
}
