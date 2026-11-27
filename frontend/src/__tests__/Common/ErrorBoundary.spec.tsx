import '../../i18n/config.ts';

import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import ErrorBoundary from '../../components/common/ErrorBoundary';

let shouldThrow = true;

function Broken() {
  if (shouldThrow) {
    throw new Error('Broken component');
  }
  return <p>Working component</p>;
}

describe('ErrorBoundary', () => {
  beforeEach(() => {
    shouldThrow = true;
    // React logs every caught render error; keep the test output clean.
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders its children when nothing throws', () => {
    shouldThrow = false;
    render(
      <ErrorBoundary>
        <Broken />
      </ErrorBoundary>
    );

    expect(screen.getByText('Working component')).toBeTruthy();
    expect(screen.queryByRole('alert')).toBeNull();
  });

  it('shows the error instead of a blank page when a child throws', () => {
    render(
      <ErrorBoundary>
        <Broken />
      </ErrorBoundary>
    );

    expect(screen.getByRole('alert')).toBeTruthy();
    expect(screen.getByText('Something went wrong')).toBeTruthy();
    expect(screen.getByText('Broken component')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Reload page' })).toBeTruthy();
  });

  it('renders the children again after "Try again" once the error is gone', () => {
    render(
      <ErrorBoundary>
        <Broken />
      </ErrorBoundary>
    );
    expect(screen.getByRole('alert')).toBeTruthy();

    shouldThrow = false;
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));

    expect(screen.getByText('Working component')).toBeTruthy();
    expect(screen.queryByRole('alert')).toBeNull();
  });

  it('only replaces the part of the page inside the boundary', () => {
    render(
      <>
        <header>Header</header>
        <ErrorBoundary>
          <Broken />
        </ErrorBoundary>
      </>
    );

    expect(screen.getByText('Header')).toBeTruthy();
    expect(screen.getByRole('alert')).toBeTruthy();
  });
});
