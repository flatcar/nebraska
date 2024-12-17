import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { fireEvent, render } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import ModalButton from '../../components/common/ModalButton';
import { theme } from '../../TestHelpers/theme';

describe('Modal Button', () => {
  it('should render Application Edit Dialog on Add Icon click', () => {
    const { getByTestId } = render(
      <MemoryRouter initialEntries={['/app/123']}>
        <ModalButton data={{}} modalToOpen="AddApplicationModal" />
      </MemoryRouter>
    );
    fireEvent.click(getByTestId('modal-button'));
    expect(getByTestId('app-edit-form')).toBeInTheDocument();
  });
  it('should render AddGroupModal on Add Icon click', () => {
    const { getByTestId } = render(
      <MemoryRouter initialEntries={['/app/123/groups/321']}>
        <ModalButton data={{}} modalToOpen="AddGroupModal" />
      </MemoryRouter>
    );
    fireEvent.click(getByTestId('modal-button'));
    expect(getByTestId('group-edit-form')).toBeInTheDocument();
  });
  it('should render AddChannelModal on Add Icon click', () => {
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
    fireEvent.click(getByTestId('modal-button'));
    expect(getByTestId('channel-edit-form')).toBeInTheDocument();
  });
  it('should render AddPackageModal on Add Icon click', () => {
    const { getByTestId } = render(
      <MemoryRouter initialEntries={['/app/123']}>
        <ModalButton data={{}} modalToOpen="AddPackageModal" />
      </MemoryRouter>
    );
    fireEvent.click(getByTestId('modal-button'));
    expect(getByTestId('package-edit-form')).toBeInTheDocument();
  });
});
