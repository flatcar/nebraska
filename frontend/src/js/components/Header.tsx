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
  menuAnchorEl: MenuProps['anchorEl'];
  projectLogo: object;
  handleClose: MenuProps['onClose'];
  handleMenu: React.MouseEventHandler<HTMLButtonElement>;
}

function Appbar(props: AppbarProps) {
  const { config, menuAnchorEl, projectLogo, handleClose, handleMenu } = props;
  const classes = useStyles();

  React.useEffect(() => {
    document.title = (config?.title) || 'Nebraska';
  }, [config]);

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
        {config?.access_management_url && (
          <IconButton
            aria-label="User menu"
            aria-controls="menu-appbar"
            aria-haspopup="true"
            onClick={handleMenu}
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
            <Button
              component="a"
              startIcon={<CreateOutlined />}
              href={config.access_management_url}
            >
              Manage Access
            </Button>
          </Menu>
        }
      </Toolbar>
    </AppBar>
  );
}

export default function Header() {
  const config = useTypedSelector(state => state.config);
  const theme = useTheme();
  const projectLogo = _.isEmpty(nebraskaLogo) ? null : nebraskaLogo;
  const [menuAnchorEl, setMenuAnchorEl] = React.useState<HTMLButtonElement | null>(null);

  function handleMenu(event: React.MouseEvent<HTMLButtonElement>) {
    setMenuAnchorEl(event.currentTarget);
  }

  function handleClose() {
    setMenuAnchorEl(null);
  }

  const props = {config, menuAnchorEl, projectLogo: projectLogo as object, handleClose, handleMenu} as AppbarProps;
  const appBar = <Appbar {...props} />
  // cachedConfig.appBarColor is for backward compatibility (the name used for the setting before).
  return config && (config.header_style === 'dark' || config.header_style === undefined && config.appBarColor === 'dark') ? (
    <ThemeProvider theme={prepareDarkTheme(theme)}>{appBar}</ThemeProvider>
  ) : (
    appBar
  );
}
