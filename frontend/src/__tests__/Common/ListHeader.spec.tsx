import '../../i18n/config.ts';

import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import ListHeader from '../../components/common/ListHeader';
import ModalButton from '../../components/common/ModalButton';
import themes from '../../lib/themes';

describe('List Header', () => {
  const minProps = {
    title: 'Applications',
    actions: [<ModalButton modalToOpen="AddApplicationModal" data={{ applications: [] }} />],
  };
  it('should render correct List Header title', () => {
    const { getByText } = render(
      <StyledEngineProvider injectFirst>
        <ThemeProvider theme={themes['light']}>
          <ListHeader title={minProps.title} />
        </ThemeProvider>
      </StyledEngineProvider>
    );
    expect(getByText(minProps.title)).toBeTruthy();
  });
  it('should render correct List Header actions', () => {
    const { asFragment } = render(
      <StyledEngineProvider injectFirst>
        <ThemeProvider theme={themes['light']}>
          <ListHeader title={minProps.title} actions={minProps.actions} />
        </ThemeProvider>
      </StyledEngineProvider>
    );
    expect(asFragment()).toMatchSnapshot();
  });
});
