import './i18n/config';

import React from 'react';
import { Provider } from 'react-redux';
import { Route, Routes } from 'react-router';

import API from './api/API';
import LoadingPage from './components/LoadingPage';
import Main from './components/Main';
import { setConfig } from './stores/redux/features/config';
import store from './stores/redux/store';

const AppRoutes = function () {
  const [configLoaded, setConfigLoaded] = React.useState(false);

  React.useEffect(() => {
    API.getConfig()
      .then(config => {
        store.dispatch(setConfig(config));
        setConfigLoaded(true);
      })
      .catch(error => {
        console.error('Failed to load config:', error);
        // Still set configLoaded to true to avoid infinite loading
        setConfigLoaded(true);
      });
  }, []);

  if (!configLoaded) {
    return <LoadingPage />;
  }

  return (
    <Provider store={store}>
      <Routes>
        <Route path="*" element={<Main />} />
      </Routes>
    </Provider>
  );
};

export default AppRoutes;
