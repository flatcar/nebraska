import { ThemeProvider } from '@mui/material/styles';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import ListView from '../../components/Instances/List';
import themes from '../../lib/themes';

describe('ListView Component', () => {
  it('renders without crashing', () => {
    const mockApplication = { id: '1', name: 'App Name' };
    const mockGroup = { id: '1', name: 'Group Name' };

    render(
      <BrowserRouter>
        <ThemeProvider theme={themes['light']}>
          <ListView application={mockApplication} group={mockGroup} />
        </ThemeProvider>
      </BrowserRouter>
    );

    expect(screen.getByText('Group Name')).toBeTruthy();
  });
});
