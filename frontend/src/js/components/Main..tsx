import green from '@material-ui/core/colors/green';
import Container from '@material-ui/core/Container';
import CssBaseline from '@material-ui/core/CssBaseline';
import { createMuiTheme } from '@material-ui/core/styles';
import { ThemeProvider } from '@material-ui/styles';
import React from 'react';
import { AuthProvider, useAuth } from 'oidc-react';
import { Route, Switch, useHistory } from 'react-router-dom';
import Footer from './Footer';
import Header from './Header';
import ApplicationLayout from './Layouts/ApplicationLayout';
import GroupLayout from './Layouts/GroupLayout';
import InstanceLayout from './Layouts/InstanceLayout';
import InstanceListLayout from './Layouts/InstanceListLayout';
import MainLayout from './Layouts/MainLayout';
import Loader from './Common/Loader';

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

const oidcConfig = {
  onSignIn: async () => {
    window.location.hash = '';
  },
  authority: process.env.REACT_APP_AUTH_AUTHORITY,
  clientId: process.env.REACT_APP_AUTH_CLIENT_ID,
  redirectUri: process.env.REACT_APP_AUTH_REDIRECT_URI,
};

function RedirectRoute() {
  const history = useHistory();
  React.useEffect(() => {
    history.replace('/');
  }, []);

  return <div>Logging You In....</div>;
}

function AppRoutes() {
  const auth = useAuth();

  if (auth?.userData?.id_token) {
    localStorage.setItem('nebraska_auth_token', auth.userData.id_token);
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
            <Route path="*" component={RedirectRoute} />
          </Switch>
          <Footer />
        </Container>
      </ThemeProvider>
    );
  }
  return <Loader />;
}

export default function Main() {
  return (
    <AuthProvider {...oidcConfig}>
      <AppRoutes />
    </AuthProvider>
  );
}
