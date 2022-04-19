import green from '@material-ui/core/colors/green';
import Container from '@material-ui/core/Container';
import CssBaseline from '@material-ui/core/CssBaseline';
import Link from '@material-ui/core/Link';
import { createMuiTheme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/core/styles';
import React from 'react';
import { Route, Switch } from 'react-router-dom';
import API from '../api/API';
import ThemeProviderNexti18n from '../i18n/ThemeProviderNexti18n';
import { setConfig } from '../stores/redux/features/config';
import { useDispatch } from '../stores/redux/hooks';
import { useAuthRedirect } from '../utils/auth';
import Footer from './Footer';
import Header from './Header';
import ApplicationLayout from './Layouts/ApplicationLayout';
import GroupLayout from './Layouts/GroupLayout';
import InstanceLayout from './Layouts/InstanceLayout';
import InstanceListLayout from './Layouts/InstanceListLayout';
import MainLayout from './Layouts/MainLayout';
import PageNotFoundLayout from './Layouts/PageNotFoundLayout';
declare module '@material-ui/core/styles/createPalette' {
  interface Palette {
    titleColor: '#000000';
    lightSilverShade: '#F0F0F0';
    greyShadeColor: '#474747';
    sapphireColor: '#061751';
  }
}

const nebraskaTheme = createMuiTheme({
  palette: {
    primary: {
      contrastText: '#fff',
      main: process.env.REACT_APP_PRIMARY_COLOR ? process.env.REACT_APP_PRIMARY_COLOR : '#2C98F0',
    },
    success: {
      main: green['500'],
      ...green,
    },
  },
  typography: {
    fontFamily: 'Overpass, sans-serif',
    h1: {
      fontSize: '1.875rem',
      fontWeight: 900,
    },
    h2: {
      fontSize: '1.875rem',
      fontWeight: 900,
    },
    h3: {
      fontSize: '1.875rem',
      fontWeight: 900,
    },
    h4: {
      fontSize: '1.875rem',
      fontWeight: 900,
    },
    subtitle1: {
      fontSize: '0.875rem',
      color: 'rgba(0,0,0,0.6)',
    },
  },
  shape: {
    borderRadius: 0,
  },
});

const useStyle = makeStyles(() => ({
  // importing visuallyHidden has typing issues at time of writing.
  // import { visuallyHidden } from '@material-ui/utils';
  visuallyHidden: {
    border: 0,
    clip: 'rect(0 0 0 0)',
    height: '1px',
    margin: -1,
    overflow: 'hidden',
    padding: 0,
    position: 'absolute',
    whiteSpace: 'nowrap',
    width: '1px',
  },
}));

export default function Main() {
  const dispatch = useDispatch();
  const classes = useStyle();

  React.useEffect(() => {
    API.getConfig().then(config => {
      console.debug('Got config', config);
      dispatch(setConfig(config));
    });
  }, []);

  useAuthRedirect();

  return (
    <ThemeProviderNexti18n theme={nebraskaTheme}>
      <CssBaseline />
      <Link href="#main" className={classes.visuallyHidden}>
        Skip to main content
      </Link>

      <Header />
      <Container component="main" id="main">
        <Switch>
          <Route path="/" exact component={MainLayout} />
          <Route path="/apps" exact component={MainLayout} />
          <Route path="/apps/:appID" exact component={ApplicationLayout} />
          <Route path="/apps/:appID/groups/:groupID" exact component={GroupLayout} />
          <Route
            path="/apps/:appID/groups/:groupID/instances"
            exact
            component={InstanceListLayout}
          />
          <Route
            path="/apps/:appID/groups/:groupID/instances/:instanceID"
            exact
            component={InstanceLayout}
          />
          <Route path="*" component={PageNotFoundLayout} />
        </Switch>
        <Footer />
      </Container>
    </ThemeProviderNexti18n>
  );
}
