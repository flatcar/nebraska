import './i18n/config';

import { Provider } from 'react-redux';
import { Route, Routes } from 'react-router-dom';

import Main from './components/Main';
import store from './stores/redux/store';

const AppRoutes = function () {
  return (
    <Provider store={store}>
      <Routes>
        <Route path="*" element={<Main />} />;
      </Routes>
    </Provider>
  );
};

export default AppRoutes;
