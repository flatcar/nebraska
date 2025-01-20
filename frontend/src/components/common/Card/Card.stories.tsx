import { Meta, Story } from '@storybook/react';
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

const CardHeaderTemplate: Story<CardHeaderProps> = args => (
  <MemoryRouter>
    <CardHeader {...args} />
  </MemoryRouter>
);
export const Header = CardHeaderTemplate.bind({});
Header.args = {
  cardDescription: 'Card Header description',
  cardMainLinkPath: '/apps/123',
  cardMainLinkLabel: 'Flatcar',
  children: 'CardHeader children',
  cardTrack: 'a cardTrack',
};

export const HeaderWithDescription = CardHeaderTemplate.bind({});
HeaderWithDescription.args = {
  cardDescription: 'Card Header description',
};
export const HeaderWithLink = CardHeaderTemplate.bind({});
HeaderWithLink.args = {
  cardDescription: 'Card Header With Link description',
  cardMainLinkPath: '/apps/123',
  cardMainLinkLabel: 'Flatcar',
};
export const HeaderWithChildren = CardHeaderTemplate.bind({});
HeaderWithChildren.args = {
  cardDescription: 'Card Header With Children description',
  children: 'CardHeader children',
};
export const HeaderWithCardTrack = CardHeaderTemplate.bind({});
HeaderWithCardTrack.args = {
  cardDescription: 'Card Header With cardTrack description',
  cardTrack: 'a cardTrack',
};
export const HeaderWithCardId = CardHeaderTemplate.bind({});
HeaderWithCardId.args = {
  cardDescription: 'Card Header With cardId description',
  cardId: 'a cardId',
};
