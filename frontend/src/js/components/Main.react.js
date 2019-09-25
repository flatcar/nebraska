import React from "react"
import {Switch, Route } from "react-router-dom"
import Header from "./Header.react"
import MainLayout from "./Layouts/MainLayout.react"
import ApplicationLayout from "./Layouts/ApplicationLayout.react"
import InstanceLayout from "./Layouts/InstanceLayout"
import GroupLayout from "./Layouts/GroupLayout.react"
import CssBaseline from '@material-ui/core/CssBaseline';
import Container from '@material-ui/core/Container';
import green from '@material-ui/core/colors/green';
import { ThemeProvider } from '@material-ui/styles';
import { createMuiTheme } from '@material-ui/core/styles';


const nebraskaTheme = createMuiTheme({
  palette: {
    primary: {
      contrastText: "#fff",
      main: process.env.PRIMARY_COLOR,
    },
    success: {
      main: green['500'],
      ...green
    },
  },
  typography: {
    fontFamily: 'Overpass, sans-serif',
  },
});

export default function Main() {
  React.useEffect(() => {
    document.title = process.env.PROJECT_NAME;
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
          <Route path="/apps/:appID/groups/:groupID/instances/:instanceID" exact component={InstanceLayout} />
        </Switch>
      </Container>
    </ThemeProvider>
  );
}
