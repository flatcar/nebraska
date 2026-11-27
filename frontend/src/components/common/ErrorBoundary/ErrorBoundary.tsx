import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import React from 'react';
import { useTranslation } from 'react-i18next';

interface ErrorFallbackProps {
  error: unknown;
  onRetry: () => void;
}

function ErrorFallback({ error, onRetry }: ErrorFallbackProps) {
  const { t } = useTranslation();
  const message = error instanceof Error ? error.message : String(error);

  return (
    <Box role="alert" padding={2} textAlign="center">
      <Typography variant="h1" style={{ fontSize: '1.5rem' }} gutterBottom>
        {t('common|error_boundary_title')}
      </Typography>
      <Typography color="textSecondary" gutterBottom>
        {t('common|error_boundary_message')}
      </Typography>
      {message && (
        <Typography
          component="pre"
          variant="body2"
          sx={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', fontFamily: 'monospace' }}
        >
          {message}
        </Typography>
      )}
      <Box display="flex" justifyContent="center" gap={1} marginTop={2}>
        <Button variant="outlined" onClick={onRetry}>
          {t('common|try_again')}
        </Button>
        <Button variant="contained" onClick={() => window.location.reload()}>
          {t('common|reload_page')}
        </Button>
      </Box>
    </Box>
  );
}

interface ErrorBoundaryProps {
  children: React.ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: unknown;
}

/**
 * Catches errors thrown while rendering its children and shows a message
 * with the error instead of leaving the page blank. React only supports
 * error boundaries as class components.
 */
export default class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { hasError: false, error: null };

  static getDerivedStateFromError(error: unknown): ErrorBoundaryState {
    return { hasError: true, error };
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null });
  };

  render() {
    if (this.state.hasError) {
      return <ErrorFallback error={this.state.error} onRetry={this.handleRetry} />;
    }

    return this.props.children;
  }
}
