import green from '@material-ui/core/colors/green';
import Container from '@material-ui/core/Container';
import CssBaseline from '@material-ui/core/CssBaseline';
import { createMuiTheme } from '@material-ui/core/styles';
import { ThemeProvider } from '@material-ui/styles';
import React from 'react';
import { useDispatch } from 'react-redux';
import { Route, Switch } from 'react-router-dom';
import API from '../api/API';
import { setConfig } from '../stores/redux/actions';
import { useAuthRedirect } from '../utils/auth';
import Footer from './Footer';
import Header from './Header';
import ApplicationLayout from './Layouts/ApplicationLayout';
import GroupLayout from './Layouts/GroupLayout';
import InstanceLayout from './Layouts/InstanceLayout';
import InstanceListLayout from './Layouts/InstanceListLayout';
import MainLayout from './Layouts/MainLayout';

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
    h4: {
      fontSize: '1.875rem',
      fontWeight: 900,
    },
    subtitle1: {
      fontSize: '14px',
    },
  },
  shape: {
    borderRadius: 0,
  },
});

export default function Main() {
  const dispatch = useDispatch();

  React.useEffect(() => {
    API.getConfig().then(config => {
      console.debug("Got config", config)
      dispatch(setConfig(config));
    });
  }, [])

  useAuthRedirect();

  return (
    <ThemeProvider theme={nebraskaTheme}>
      <CssBaseline />
      <Header />
      <Container component="main">
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
        </Switch>
        <Footer />
      </Container>
    </ThemeProvider>
  );
}
