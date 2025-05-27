import { createStore } from '@reduxjs/toolkit';
import { Meta, StoryFn } from '@storybook/react';
import { Provider } from 'react-redux';
import { MemoryRouter } from 'react-router';

import FooterComponent from './Footer';

export default {
  title: 'Footer',
  component: FooterComponent,
  argTypes: {},
  decorators: [
    Story => {
      return (
        <MemoryRouter>
          <Story />
        </MemoryRouter>
      );
    },
  ],
} as Meta;

const Template: StoryFn = args => {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const store = createStore((state = { config: {} }, _action) => state, {
    config: {
      ...args,
    },
  });
  return (
    <Provider store={store}>
      <FooterComponent />
    </Provider>
  );
};

export const FooterNoOverride = {
  render: Template,

  args: {
    title: '',
    nebraska_version: '',
  },
};

export const FooterOverride = {
  render: Template,

  args: {
    title: 'Some Pro Update Service',
    nebraska_version: '1.2.3',
  },
};
