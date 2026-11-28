import React from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router';

import AppRoutes from './App';
import ErrorBoundary from './components/common/ErrorBoundary';

if (import.meta.env.MODE !== 'production') {
  import('@axe-core/react').then(axe => {
    axe.default(React, createRoot, 1000);
  });
}

const root = createRoot(document.getElementById('root')!);
root.render(
  <BrowserRouter>
    {/* Last-resort boundary for errors outside the page content, e.g. in the header */}
    <ErrorBoundary>
      <AppRoutes />
    </ErrorBoundary>
  </BrowserRouter>
);
