import { Icon } from '@iconify/react';
import { Box, Link } from '@material-ui/core';
import AppBar from '@material-ui/core/AppBar';
import IconButton from '@material-ui/core/IconButton';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import Menu, { MenuProps } from '@material-ui/core/Menu';
import { createMuiTheme, makeStyles, Theme, ThemeProvider, useTheme } from '@material-ui/core/styles';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import AccountCircle from '@material-ui/icons/AccountCircle';
import CreateOutlined from '@material-ui/icons/CreateOutlined';
import DOMPurify from 'dompurify';
import React from 'react';
import _ from 'underscore';
import API from '../api/API';
import nebraskaLogo from '../icons/nebraska-logo.json';

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
  cachedConfig: NebraskaConfig;
  menuAnchorEl: MenuProps['anchorEl'];
  projectLogo: object;
  config: NebraskaConfig | null;
  handleClose: MenuProps['onClose'];
  handleMenu: React.MouseEventHandler<HTMLButtonElement>;
}

function Appbar(props: AppbarProps) {
  const { cachedConfig, menuAnchorEl, projectLogo, config, handleClose, handleMenu } = props;
  const classes = useStyles();

  React.useEffect(() => {
    document.title = (cachedConfig && cachedConfig.title) || 'Nebraska';
  }, [cachedConfig]);

  return (
    <AppBar position="static" className={classes.header}>
      <Toolbar>
        {cachedConfig?.logo ? (
          <Box className={classes.svgContainer}>
            <div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(cachedConfig.logo) }} />
          </Box>
        ) : (
          <Icon icon={projectLogo} height={45} />
        )}
        {cachedConfig && cachedConfig.title !== '' && (
          <Typography variant="h6" className={classes.title}>
            {cachedConfig.title}
          </Typography>
        )}

        {config?.access_management_url && (
          <IconButton
            aria-label="User menu"
            aria-controls="menu-appbar"
            aria-haspopup="true"
            onClick={handleMenu}
            color="inherit"
          >
            <AccountCircle />
          </IconButton>
        )}
        {config?.access_management_url &&
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
            <Link href={config.access_management_url}>
              <ListItemIcon>
                <CreateOutlined />
              </ListItemIcon>
              <ListItemText primary="Manage Access" />
            </Link>
          </Menu>
        }
      </Toolbar>
    </AppBar>
  );
}

export default function Header() {
  const [config, setConfig] = React.useState<NebraskaConfig | null>(null);
  const theme = useTheme();
  const projectLogo = _.isEmpty(nebraskaLogo) ? null : nebraskaLogo;

  const [menuAnchorEl, setMenuAnchorEl] = React.useState<HTMLButtonElement | null>(null);
  const [cachedConfig, setCachedConfig] = React.useState<NebraskaConfig>(
    JSON.parse(localStorage.getItem('nebraska_config') || "") as NebraskaConfig
  );

  function handleMenu(event: React.MouseEvent<HTMLButtonElement>) {
    setMenuAnchorEl(event.currentTarget);
  }

  function handleClose() {
    setMenuAnchorEl(null);
  }

  // @todo: This should be abstracted but we should do it when we integrate Redux.
  React.useEffect(() => {
    if (!config) {
      API.getConfig()
        .then(config => {
          const cacheConfig: NebraskaConfig = {
            title: config.title,
            logo: config.logo,
            header_style: config.header_style,
          };
          localStorage.setItem('nebraska_config', JSON.stringify(cacheConfig));
          setCachedConfig(cacheConfig);
          setConfig(config);
        })
        .catch(error => {
          console.error(error);
        });
    }
  }, [config]);

  const props = {cachedConfig, menuAnchorEl, projectLogo: projectLogo as object, config, handleClose, handleMenu} as AppbarProps;
  const appBar = <Appbar {...props} />
  // cachedConfig.appBarColor is for backward compatibility (the name used for the setting before).
  return cachedConfig && (cachedConfig.header_style === 'dark' || cachedConfig.header_style === undefined && cachedConfig.appBarColor === 'dark') ? (
    <ThemeProvider theme={prepareDarkTheme(theme)}>{appBar}</ThemeProvider>
  ) : (
    appBar
  );
}
