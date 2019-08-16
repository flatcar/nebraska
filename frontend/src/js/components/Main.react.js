import React from "react"
import {Switch, Route } from "react-router-dom"
import Header from "./Header.react"
import ProgressBar from "./ProgressBar.react"
import MainLayout from "./Layouts/MainLayout.react"
import ApplicationLayout from "./Layouts/ApplicationLayout.react"
import GroupLayout from "./Layouts/GroupLayout.react"


export default function Main() {
  return (
    <Header />
    <ProgressBar name="main_progress_bar" color="#ddd" width={0.2} />
    <Switch>
      <Route path='/' exact component={MainLayout} />
      <Route path='/apps' exact component={MainLayout} />
      <Route path="/apps/:appID" exact component={ApplicationLayout} />
      <Route path="/apps/:appID/groups/:groupID" component={GroupLayout}/>
    </Switch>
  );
}
