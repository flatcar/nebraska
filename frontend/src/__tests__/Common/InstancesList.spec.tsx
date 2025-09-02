import { ThemeProvider } from '@mui/material/styles';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router';
import { describe, expect, it, vi } from 'vitest';

import { Application, Group } from '../../api/apiDataTypes';
import ListView from '../../components/Instances/List';
import themes from '../../lib/themes';

describe('ListView Component', () => {
  global.fetch = vi.fn(() =>
    Promise.resolve(
      new Response(JSON.stringify({ count: 10 }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    )
  );
  it('renders without crashing', () => {
    const mockApplication = { id: '1', name: 'App Name' } as unknown as Application;
    const mockGroup = { id: '1', name: 'Group Name' } as unknown as Group;

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
