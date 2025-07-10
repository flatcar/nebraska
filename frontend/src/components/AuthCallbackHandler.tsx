import Loader from './common/Loader/Loader';

export default function AuthCallbackHandler() {
  // This component simply shows a loader while the auth callback is being processed
  // The actual authentication handling is done by the useAuthRedirect hook in Main.tsx
  return <Loader />;
}
