import Container from '@mui/material/Container';
import CssBaseline from '@mui/material/CssBaseline';
import Link from '@mui/material/Link';
import { visuallyHidden } from '@mui/utils';
import { Outlet, Route, Routes } from 'react-router';

import ThemeProviderNexti18n from '../i18n/ThemeProviderNexti18n';
import themes, { getThemeName, usePrefersColorScheme } from '../lib/themes';
import AuthGuard from './AuthGuard';
import Footer from './Footer';
import Header from './Header';
import ApplicationLayout from './layouts/ApplicationLayout';
import AuthErrorLayout from './layouts/AuthErrorLayout';
import GroupLayout from './layouts/GroupLayout';
import InstanceLayout from './layouts/InstanceLayout';
import InstanceListLayout from './layouts/InstanceListLayout';
import MainLayout from './layouts/MainLayout';
import PageNotFoundLayout from './layouts/PageNotFoundLayout';
import LoadingPage from './LoadingPage';

function SkipLink() {
  return (
    <Link href="#main" style={visuallyHidden} underline="hover">
      Skip to main content
    </Link>
  );
}

export default function Main() {
  // let themeName = useTypedSelector(state => state.ui.theme.name);
  usePrefersColorScheme();

  const themeName = getThemeName() || 'light';

  return (
    <ThemeProviderNexti18n theme={themes[themeName]}>
      <CssBaseline />
      <SkipLink />
      <Header />
      <Container component="main" id="main" sx={{ paddingTop: '0.52rem' }}>
        <Routes>
          <Route path="/404" element={<PageNotFoundLayout />} />
          <Route path="/auth/error" element={<AuthErrorLayout />} />
          <Route
            element={
              <AuthGuard>
                <Outlet />
              </AuthGuard>
            }
          >
            <Route path="/" element={<MainLayout />} />
            <Route path="/apps" element={<MainLayout />} />
            <Route path="/apps/:appID" element={<ApplicationLayout />} />
            <Route path="/apps/:appID/groups/:groupID" element={<GroupLayout />} />
            <Route path="/apps/:appID/groups/:groupID/instances" element={<InstanceListLayout />} />
            <Route
              path="/apps/:appID/groups/:groupID/instances/:instanceID"
              element={<InstanceLayout />}
            />
            <Route path="/auth/callback" element={<LoadingPage />} />
          </Route>
          <Route path="*" element={<PageNotFoundLayout />} />
        </Routes>
        <Footer />
      </Container>
    </ThemeProviderNexti18n>
  );
}
