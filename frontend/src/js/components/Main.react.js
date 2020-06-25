import green from '@material-ui/core/colors/green';
import Container from '@material-ui/core/Container';
import CssBaseline from '@material-ui/core/CssBaseline';
import { createMuiTheme } from '@material-ui/core/styles';
import { ThemeProvider } from '@material-ui/styles';
import React from 'react';
import {Route, Switch } from 'react-router-dom';
import Header from './Header.react';
import ApplicationLayout from './Layouts/ApplicationLayout.react';
import GroupLayout from './Layouts/GroupLayout.react';
import InstanceLayout from './Layouts/InstanceLayout';
import InstanceListLayout from './Layouts/InstanceListLayout';
import MainLayout from './Layouts/MainLayout.react';

const nebraskaTheme = createMuiTheme({
  palette: {
    primary: {
      contrastText: '#fff',
      main: process.env.REACT_APP_PRIMARY_COLOR,
    },
    success: {
      main: green['500'],
      ...green
    },
    titleColor: '#000000',
    lightSilverShade: '#F0F0F0',
    greyShadeColor: '#474747'
  },
  typography: {
    fontFamily: 'Overpass, sans-serif',
    h4: {
      fontSize: '1.875rem',
      fontWeight: 900
    },
    subtitle1: {
      fontSize: '14px'
    }
  },
  shape: {
    borderRadius: 0
  }
});

export default function Main() {
  React.useEffect(() => {
    document.title = process.env.REACT_APP_PROJECT_NAME;
  },
  []);

  return (
    <ThemeProvider theme={nebraskaTheme}>
      <CssBaseline />
      <Header />
      <Container component="main">
        <Switch>
          <Route path='/' exact component={MainLayout} />
          <Route path='/apps' exact component={MainLayout} />
          <Route path="/apps/:appID" exact component={ApplicationLayout} />
          <Route path="/apps/:appID/groups/:groupID" exact component={GroupLayout} />
          <Route path="/apps/:appID/groups/:groupID/instances" exact component={InstanceListLayout} />
          <Route path="/apps/:appID/groups/:groupID/instances/:instanceID" exact component={InstanceLayout} />
        </Switch>
      </Container>
    </ThemeProvider>
  );
}
