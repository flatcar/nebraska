import React from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';

import AppRoutes from './App';

if (import.meta.env.MODE !== 'production') {
  import('@axe-core/react').then(axe => {
    axe.default(React, createRoot, 1000);
  });
}

const root = createRoot(document.getElementById('root')!);
root.render(
  <BrowserRouter>
    <AppRoutes />
  </BrowserRouter>
);
