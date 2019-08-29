import React from "react"
import {Switch, Route } from "react-router-dom"
import Header from "./Header.react"
import ProgressBar from "./ProgressBar.react"
import MainLayout from "./Layouts/MainLayout.react"
import ApplicationLayout from "./Layouts/ApplicationLayout.react"
import GroupLayout from "./Layouts/GroupLayout.react"
import Container from '@material-ui/core/Container';
import { ThemeProvider } from '@material-ui/styles';
import { createMuiTheme } from '@material-ui/core/styles';

const nebraskaTheme = createMuiTheme({
  palette: {
    primary: {
      contrastText: "#fff",
      main: '#00AEEF',
    },
  },
  typography: {
    // Account for base font-size of 62.5%.
    htmlFontSize: 10,
  },
});

export default function Main() {
  return (
    <ThemeProvider theme={nebraskaTheme}>
      <Header />
      <ProgressBar name="main_progress_bar" color="#ddd" width={0.2} />
      <Container component="main">
        <Switch>
          <Route path='/' exact component={MainLayout} />
          <Route path='/apps' exact component={MainLayout} />
          <Route path="/apps/:appID" exact component={ApplicationLayout} />
          <Route path="/apps/:appID/groups/:groupID" component={GroupLayout}/>
        </Switch>
      </Container>
    </ThemeProvider>
  );
}
