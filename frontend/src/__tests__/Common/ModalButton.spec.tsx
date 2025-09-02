import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { act, fireEvent, render } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { describe, expect, it } from 'vitest';

import ModalButton from '../../components/common/ModalButton';
import { theme } from '../../TestHelpers/theme';

describe('Modal Button', () => {
  it('should render Application Edit Dialog on Add Icon click', async () => {
    const { getByTestId } = render(
      <StyledEngineProvider injectFirst>
        <ThemeProvider theme={theme}>
          <MemoryRouter initialEntries={['/app/123']}>
            <ModalButton data={{}} modalToOpen="AddApplicationModal" />
          </MemoryRouter>
        </ThemeProvider>
      </StyledEngineProvider>
    );
    await act(async () => fireEvent.click(getByTestId('modal-button')));
    expect(getByTestId('app-edit-form')).toBeTruthy();
  });
  it('should render AddGroupModal on Add Icon click', async () => {
    const { getByTestId } = render(
      <StyledEngineProvider injectFirst>
        (
        <ThemeProvider theme={theme}>
          <MemoryRouter initialEntries={['/app/123/groups/321']}>
            <ModalButton data={{}} modalToOpen="AddGroupModal" />
          </MemoryRouter>
        </ThemeProvider>
        )
      </StyledEngineProvider>
    );
    await act(async () => fireEvent.click(getByTestId('modal-button')));
    expect(getByTestId('group-edit-form')).toBeTruthy();
  });
  it('should render AddChannelModal on Add Icon click', async () => {
    const tree = (
      <StyledEngineProvider injectFirst>
        (
        <ThemeProvider theme={theme}>
          <MemoryRouter initialEntries={['/app/123']}>
            <ModalButton data={{}} modalToOpen="AddChannelModal" />
          </MemoryRouter>
        </ThemeProvider>
        )
      </StyledEngineProvider>
    );
    const { getByTestId } = render(tree);
    await act(async () => fireEvent.click(getByTestId('modal-button')));
    expect(getByTestId('channel-edit-form')).toBeTruthy();
  });
  it('should render AddPackageModal on Add Icon click', async () => {
    const { getByTestId } = render(
      <StyledEngineProvider injectFirst>
        (
        <ThemeProvider theme={theme}>
          <MemoryRouter initialEntries={['/app/123']}>
            <ModalButton data={{}} modalToOpen="AddPackageModal" />
          </MemoryRouter>
        </ThemeProvider>
        )
      </StyledEngineProvider>
    );
    await act(async () => fireEvent.click(getByTestId('modal-button')));
    expect(getByTestId('package-edit-form')).toBeTruthy();
  });
});
