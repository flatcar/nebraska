import { Icon, IconifyIcon } from '@iconify/react';
import AccountCircle from '@mui/icons-material/AccountCircle';
import CreateOutlined from '@mui/icons-material/CreateOutlined';
import LogoutOutlined from '@mui/icons-material/LogoutOutlined';
import { Box, Button, Divider } from '@mui/material';
import AppBar from '@mui/material/AppBar';
import IconButton from '@mui/material/IconButton';
import Menu, { MenuProps } from '@mui/material/Menu';
import { styled } from '@mui/material/styles';
import { StyledEngineProvider, Theme, ThemeProvider } from '@mui/material/styles';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import DOMPurify from 'dompurify';
import React from 'react';
import { useTranslation } from 'react-i18next';
import _ from 'underscore';

import nebraskaLogo from '../icons/nebraska-logo.json';
import themes from '../lib/themes';
import { setUser, UserState } from '../stores/redux/features/user';
import { useDispatch, useSelector } from '../stores/redux/hooks';
import { broadcastLogout, getIdToken } from '../utils/auth';
import { getOIDCClient } from '../utils/oidc';

const PREFIX = 'Header';

const classes = {
  title: `${PREFIX}-title`,
  header: `${PREFIX}-header`,
  svgContainer: `${PREFIX}-svgContainer`,
  userName: `${PREFIX}-userName`,
  email: `${PREFIX}-email`,
};

const StyledStyledEngineProvider = styled(StyledEngineProvider)(({ theme }) => ({
  [`& .${classes.title}`]: {
    flexGrow: 1,
    display: 'none',
    color: theme.palette.titleColor,
    [theme.breakpoints.up('sm')]: {
      display: 'block',
    },
  },

  [`& .${classes.header}`]: {
    marginBottom: theme.spacing(1),
  },

  [`& .${classes.svgContainer}`]: {
    '& svg': { maxHeight: '3rem' },
  },

  [`& .${classes.userName}`]: {
    fontSize: '1.5em',
  },

  [`& .${classes.email}`]: {
    fontSize: '1.2em',
  },
}));

declare module '@mui/material/styles' {
  // eslint-disable-next-line @typescript-eslint/no-empty-object-type
  interface DefaultTheme extends Theme {}
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
  projectLogo: IconifyIcon;
  handleClose: MenuProps['onClose'];
  handleMenu: React.MouseEventHandler<HTMLButtonElement>;
  handleLogout: () => void;
}

function Appbar(props: AppbarProps) {
  const { config, user, menuAnchorEl, projectLogo, handleClose, handleMenu, handleLogout } = props;

  const { t } = useTranslation();

  React.useEffect(() => {
    document.title = config?.title || 'Nebraska';
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
            aria-label={t('header|user_menu')}
            aria-controls="menu-appbar"
            aria-haspopup="true"
            onClick={handleMenu}
            size="large"
          >
            <AccountCircle />
          </IconButton>
        )}
        {showAccountButton && (
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
            {(user?.name || user?.email) && (
              <Box paddingY={2} paddingX={2} textAlign="center">
                {user?.name && <Typography className={classes.userName}>{user.name}</Typography>}
                {user?.email && <Typography className={classes.email}>{user.email}</Typography>}
              </Box>
            )}
            <Box paddingY={1} paddingX={2} textAlign="center">
              <Button
                component="a"
                startIcon={<CreateOutlined />}
                disabled={!config?.access_management_url}
                href={config?.access_management_url || ''}
              >
                {t('header|manage_account')}
              </Button>
            </Box>
            {user?.authenticated && config?.auth_mode === 'oidc' && (
              <>
                <Divider />
                <Box paddingY={1} paddingX={2} textAlign="center">
                  <Button startIcon={<LogoutOutlined />} onClick={handleLogout} fullWidth>
                    {t('header|logout')}
                  </Button>
                </Box>
              </>
            )}
          </Menu>
        )}
      </Toolbar>
    </AppBar>
  );
}

export default function Header() {
  const { config, user } = useSelector(state => ({ config: state.config, user: state.user }));
  const dispatch = useDispatch();
  const projectLogo = _.isEmpty(nebraskaLogo) ? null : nebraskaLogo;
  const [menuAnchorEl, setMenuAnchorEl] = React.useState<HTMLButtonElement | null>(null);

  function handleMenu(event: React.MouseEvent<HTMLButtonElement>) {
    setMenuAnchorEl(event.currentTarget);
  }

  function handleClose() {
    setMenuAnchorEl(null);
  }

  function handleLogout() {
    // Use OIDC client for logout (direct provider logout only)
    const oidcClient = getOIDCClient();
    if (oidcClient && config?.auth_mode === 'oidc') {
      const logoutRedirectUrl = new URL(window.location.href);
      logoutRedirectUrl.pathname = '/';
      // Get ID token BEFORE clearing tokens
      const idTokenHint = getIdToken();

      // Clear tokens and user state
      broadcastLogout();
      dispatch(setUser({ authenticated: false, name: '', email: '' }));

      // Redirect to OIDC logout endpoint
      oidcClient.logout(logoutRedirectUrl.toString(), idTokenHint || undefined);
    } else {
      // For non-OIDC auth modes, just clear tokens and user state
      broadcastLogout();
      dispatch(setUser({ authenticated: false, name: '', email: '' }));
    }
  }

  const props = {
    config,
    user,
    menuAnchorEl,
    projectLogo: projectLogo as object,
    handleClose,
    handleMenu,
    handleLogout,
  } as AppbarProps;
  const appBar = <Appbar {...props} />;
  // cachedConfig.appBarColor is for backward compatibility (the name used for the setting before).
  // @todo: Use themes@getThemeName to get the name.
  return config &&
    (config.header_style === 'dark' ||
      (config.header_style === undefined && config.appBarColor === 'dark')) ? (
    <StyledStyledEngineProvider injectFirst>
      (<ThemeProvider theme={themes.dark}>{appBar}</ThemeProvider>)
    </StyledStyledEngineProvider>
  ) : (
    appBar
  );
}
