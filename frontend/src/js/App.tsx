import React from 'react';
import { Provider } from "react-redux";
import { Route } from 'react-router-dom';
import Main from './components/Main.';
import store from "./stores/redux/store";

var AppRoutes = function () {
  return (
    <Provider store={store}>
      <Route path="/" component={Main} />;
    </Provider>
  );
};

export default AppRoutes;
