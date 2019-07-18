import React from "react"
import ReactDOM from "react-dom"
import { HashRouter as Router, Route } from "react-router-dom"
import Main from "./components/Main.react"

var routes = (
  <Route path='/' component={Main} />
)

ReactDOM.render(<Router>{routes}</Router>, document.getElementById('root'))
