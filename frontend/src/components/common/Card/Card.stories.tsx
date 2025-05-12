import { Meta, StoryFn } from '@storybook/react';
import { MemoryRouter } from 'react-router-dom';

import {
  CardDescriptionLabel,
  CardFeatureLabel,
  CardHeader,
  CardHeaderProps,
  CardLabel,
} from './Card';

export default {
  title: 'Card',
} as Meta;

export const FeatureLabel = () => <CardFeatureLabel>Feature Label</CardFeatureLabel>;

export const DescriptionLabel = () => (
  <CardDescriptionLabel>Description Label</CardDescriptionLabel>
);

export const Label = () => <CardLabel>Label</CardLabel>;
export const LabelWithStyles = () => <CardLabel labelStyle={{}}>Label</CardLabel>;

const CardHeaderTemplate: StoryFn<CardHeaderProps> = args => (
  <MemoryRouter>
    <CardHeader {...args} />
  </MemoryRouter>
);

export const Header = {
  render: CardHeaderTemplate,

  args: {
    cardDescription: 'Card Header description',
    cardMainLinkPath: '/apps/123',
    cardMainLinkLabel: 'Flatcar',
    children: 'CardHeader children',
    cardTrack: 'a cardTrack',
  },
};

export const HeaderWithDescription = {
  render: CardHeaderTemplate,

  args: {
    cardDescription: 'Card Header description',
  },
};

export const HeaderWithLink = {
  render: CardHeaderTemplate,

  args: {
    cardDescription: 'Card Header With Link description',
    cardMainLinkPath: '/apps/123',
    cardMainLinkLabel: 'Flatcar',
  },
};

export const HeaderWithChildren = {
  render: CardHeaderTemplate,

  args: {
    cardDescription: 'Card Header With Children description',
    children: 'CardHeader children',
  },
};

export const HeaderWithCardTrack = {
  render: CardHeaderTemplate,

  args: {
    cardDescription: 'Card Header With cardTrack description',
    cardTrack: 'a cardTrack',
  },
};

export const HeaderWithCardId = {
  render: CardHeaderTemplate,

  args: {
    cardDescription: 'Card Header With cardId description',
    cardId: 'a cardId',
  },
};
