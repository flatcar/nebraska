import React from 'react';
import { Route } from 'react-router-dom';
import Main from './components/Main.';

var AppRoutes = function() {
  return (<Route path='/' component={Main} />);
};

export default AppRoutes;
