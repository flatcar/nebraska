import React from 'react';

import { useSelector } from '../stores/redux/hooks';
import { useAuthRedirect, useLogoutSync } from '../utils/auth';
import LoadingPage from './LoadingPage';

interface AuthGuardProps {
  children: React.ReactNode;
}

export default function AuthGuard({ children }: AuthGuardProps) {
  const { authLoading, authenticated } = useSelector(state => state.user);
  const config = useSelector(state => state.config);

  // Handle authentication
  useAuthRedirect();
  useLogoutSync();

  // Show loader while auth state is being determined or if not authenticated in OIDC mode
  if (authLoading || (config.auth_mode === 'oidc' && !authenticated)) {
    return <LoadingPage />;
  }

  // Auth state is known, render children
  return children;
}
