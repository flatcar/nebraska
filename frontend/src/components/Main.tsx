import Container from '@material-ui/core/Container';
import CssBaseline from '@material-ui/core/CssBaseline';
import Link from '@material-ui/core/Link';
import { makeStyles } from '@material-ui/core/styles';
import React from 'react';
import { Route, Switch } from 'react-router-dom';
import API from '../api/API';
import ThemeProviderNexti18n from '../i18n/ThemeProviderNexti18n';
import themes, { getThemeName, usePrefersColorScheme } from '../lib/themes';
import { setConfig } from '../stores/redux/features/config';
import { useDispatch } from '../stores/redux/hooks';
import { useAuthRedirect } from '../utils/auth';
import Footer from './Footer';
import Header from './Header';
import ApplicationLayout from './layouts/ApplicationLayout';
import GroupLayout from './layouts/GroupLayout';
import InstanceLayout from './layouts/InstanceLayout';
import InstanceListLayout from './layouts/InstanceListLayout';
import MainLayout from './layouts/MainLayout';
import PageNotFoundLayout from './layouts/PageNotFoundLayout';

const useStyle = makeStyles(() => ({
  // importing visuallyHidden has typing issues at time of writing.
  // import { visuallyHidden } from '@material-ui/utils';
  visuallyHidden: {
    border: 0,
    clip: 'rect(0 0 0 0)',
    height: '1px',
    margin: -1,
    overflow: 'hidden',
    padding: 0,
    position: 'absolute',
    whiteSpace: 'nowrap',
    width: '1px',
  },
}));

export default function Main() {
  const dispatch = useDispatch();
  const classes = useStyle();
  // let themeName = useTypedSelector(state => state.ui.theme.name);
  let themeName = 'light';
  usePrefersColorScheme();

  if (!themeName) {
    themeName = getThemeName();
  }

  React.useEffect(() => {
    API.getConfig().then(config => {
      console.debug('Got config', config);
      dispatch(setConfig(config));
    });
  }, []);

  useAuthRedirect();

  return (
    <ThemeProviderNexti18n theme={themes[themeName]}>
      <CssBaseline />
      <Link href="#main" className={classes.visuallyHidden}>
        Skip to main content
      </Link>

      <Header />
      <Container component="main" id="main">
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
          <Route path="*" component={PageNotFoundLayout} />
        </Switch>
        <Footer />
      </Container>
    </ThemeProviderNexti18n>
  );
}
